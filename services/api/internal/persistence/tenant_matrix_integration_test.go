package persistence

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fikrimastor/mytax/services/api/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// These tests need the two-tenant fixture (testdata/security_fixture.sql) and connect as the
// restricted application role. They skip only when no database is configured; CI fails on any skip.
//
// Fixture: tenant A has subject "test-subject" and one evidence row; tenant B has subject
// "test-subject-b" and one evidence row.

const (
	testIssuer   = "https://issuer.example"
	testAudience = "mytax-api"
	subjectA     = "test-subject"
	subjectB     = "test-subject-b"
)

type fixture struct {
	url, tenantA, tenantB string
}

func needDB(t *testing.T) fixture {
	t.Helper()
	f := fixture{os.Getenv("MYTAX_TEST_DATABASE_URL"), os.Getenv("MYTAX_TEST_TENANT_A"), os.Getenv("MYTAX_TEST_TENANT_B")}
	if f.url == "" || f.tenantA == "" || f.tenantB == "" {
		t.Skip("integration DB fixture not configured")
	}
	return f
}

func newPool(t *testing.T, f fixture, maxConns int32) *pgxpool.Pool {
	t.Helper()
	cfg, err := pgxpool.ParseConfig(f.url)
	if err != nil {
		t.Fatal(err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func ident(subject string) security.Identity {
	return security.Identity{Subject: subject, Issuer: testIssuer, Audience: testAudience, Verified: true}
}

// rawTx runs fn in a read-write transaction, optionally with the transaction-local subject set,
// and ALWAYS rolls back so the fixture is never changed.
func rawTx(ctx context.Context, pool *pgxpool.Pool, subject string, setSubject bool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if setSubject {
		if _, err := tx.Exec(ctx, "select set_config('app.subject', $1, true)", subject); err != nil {
			return err
		}
	}
	return fn(tx)
}

func count(ctx context.Context, tx pgx.Tx, table string) (int, error) {
	var n int
	err := tx.QueryRow(ctx, "select count(*) from "+table).Scan(&n)
	return n, err
}

func sqlState(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return ""
}

func TestSymmetricTwoTenantIsolation(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 0)
	db, err := NewTenantDB(pool, testIssuer, testAudience)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		subject, own, other string
	}{{subjectA, f.tenantA, f.tenantB}, {subjectB, f.tenantB, f.tenantA}}
	for _, c := range cases {
		if n, err := db.CountEvidence(ctx, ident(c.subject), c.own); err != nil || n != 1 {
			t.Fatalf("%s own tenant: n=%d err=%v", c.subject, n, err)
		}
		if _, err := db.CountEvidence(ctx, ident(c.subject), c.other); err == nil || err.Error() != "tenant membership denied" {
			t.Fatalf("%s must be denied the other tenant, got %v", c.subject, err)
		}
		// The database itself must hide the other tenant, independent of the Go membership check.
		err := rawTx(ctx, pool, c.subject, true, func(tx pgx.Tx) error {
			for _, table := range []string{"tenants", "tenant_memberships", "evidence_metadata"} {
				if n, err := count(ctx, tx, table); err != nil || n != 1 {
					return fmt.Errorf("%s sees %d rows in %s (err=%v), want exactly its own 1", c.subject, n, table, err)
				}
			}
			var seen string
			if err := tx.QueryRow(ctx, "select tenant_id::text from evidence_metadata").Scan(&seen); err != nil || seen != c.own {
				return fmt.Errorf("%s sees evidence of tenant %q, want %q (err=%v)", c.subject, seen, c.own, err)
			}
			if n, err := count(ctx, tx, "audit_events"); err != nil || n != 0 {
				return fmt.Errorf("audit_events n=%d err=%v", n, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSubjectWithoutMembershipOrWithoutSubjectSeesNothing(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 0)
	for name, c := range map[string]struct {
		subject string
		set     bool
	}{
		"unknown subject":     {"nobody", true},
		"empty subject":       {"", true},
		"no subject set":      {"", false},
		"sql-looking subject": {"' OR '1'='1", true},
		"other case":          {strings.ToUpper(subjectA), true},
	} {
		err := rawTx(ctx, pool, c.subject, c.set, func(tx pgx.Tx) error {
			for _, table := range []string{"tenants", "tenant_memberships", "evidence_metadata", "audit_events"} {
				if n, err := count(ctx, tx, table); err != nil || n != 0 {
					return fmt.Errorf("%s: %s exposed %d rows (err=%v)", name, table, n, err)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRLSRejectsCrossTenantWriteAndOwnWriteIsRolledBack(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 0)
	insert := "insert into evidence_metadata(tenant_id,claim_id,sha256_hex,object_key,byte_size,content_type) values ($1::uuid, gen_random_uuid(), repeat('c',64), $2, 1, 'application/pdf')"

	// Own tenant: allowed by the WITH CHECK policy (then rolled back).
	if err := rawTx(ctx, pool, subjectA, true, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, insert, f.tenantA, "matrix-own")
		return err
	}); err != nil {
		t.Fatalf("own-tenant insert must be allowed by RLS: %v", err)
	}
	// Wrong tenant, no subject, unknown subject: RLS must refuse (SQLSTATE 42501, row-level security).
	for name, c := range map[string]struct {
		subject string
		set     bool
		tenant  string
	}{
		"A writes into B":    {subjectA, true, f.tenantB},
		"B writes into A":    {subjectB, true, f.tenantA},
		"no subject into A":  {"", false, f.tenantA},
		"unknown into A":     {"nobody", true, f.tenantA},
		"empty subject to B": {"", true, f.tenantB},
	} {
		err := rawTx(ctx, pool, c.subject, c.set, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, insert, c.tenant, "matrix-"+name)
			return err
		})
		if err == nil || sqlState(err) != "42501" || !strings.Contains(err.Error(), "row-level security") {
			t.Fatalf("%s must be refused by RLS, got %v", name, err)
		}
	}
	// The failed and rolled-back writes left the fixture unchanged.
	var rows int
	if err := rawTx(ctx, pool, subjectA, true, func(tx pgx.Tx) error {
		var e error
		rows, e = count(ctx, tx, "evidence_metadata")
		return e
	}); err != nil || rows != 1 {
		t.Fatalf("fixture changed: rows=%d err=%v", rows, err)
	}
}

func TestApplicationRoleCannotEscalateOrTamper(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 0)
	stmts := map[string]string{
		"add self to tenant B":       fmt.Sprintf("insert into tenant_memberships(tenant_id,subject_id,role) values ('%s','%s','owner')", f.tenantB, subjectA),
		"promote self":               fmt.Sprintf("update tenant_memberships set role='owner' where subject_id='%s'", subjectA),
		"remove membership":          "delete from tenant_memberships",
		"create tenant":              "insert into tenants default values",
		"edit tenant":                "update tenants set created_at=now()",
		"delete tenant":              "delete from tenants",
		"forge audit event":          fmt.Sprintf("insert into audit_events(tenant_id,subject_id,action,resource_id,event_digest) values ('%s','x','x','x',repeat('a',64))", f.tenantA),
		"edit audit event":           "update audit_events set action='x'",
		"delete audit event":         "delete from audit_events",
		"edit evidence":              "update evidence_metadata set byte_size=0",
		"delete evidence":            "delete from evidence_metadata",
		"truncate evidence":          "truncate evidence_metadata",
		"drop table":                 "drop table evidence_metadata",
		"disable row level security": "alter table evidence_metadata disable row level security",
		"turn off force rls":         "alter table evidence_metadata no force row level security",
		"bypass via role attribute":  "alter role mytax_app bypassrls",
		"set role to the owner":      "set role postgres",
	}
	for name, stmt := range stmts {
		err := rawTx(ctx, pool, subjectA, true, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, stmt)
			return err
		})
		if err == nil {
			t.Fatalf("%s must fail but succeeded: %s", name, stmt)
		}
		if code := sqlState(err); code != "42501" {
			t.Fatalf("%s must fail with insufficient_privilege 42501, got %q: %v", name, code, err)
		}
	}
	// The role flags are still safe afterwards.
	var super, bypass bool
	if err := pool.QueryRow(ctx, "select rolsuper, rolbypassrls from pg_roles where rolname=current_user").Scan(&super, &bypass); err != nil || super || bypass {
		t.Fatalf("role became unsafe: super=%v bypass=%v err=%v", super, bypass, err)
	}
}

func TestSpoofedOrMalformedIdentityAndTenantNeverReachData(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 0)
	db, _ := NewTenantDB(pool, testIssuer, testAudience)
	good := ident(subjectA)
	bad := map[string]struct {
		id     security.Identity
		tenant string
	}{
		"unverified":         {security.Identity{Subject: subjectA, Issuer: testIssuer, Audience: testAudience, Verified: false}, f.tenantA},
		"wrong issuer":       {security.Identity{Subject: subjectA, Issuer: "https://evil.example", Audience: testAudience, Verified: true}, f.tenantA},
		"wrong audience":     {security.Identity{Subject: subjectA, Issuer: testIssuer, Audience: "other-api", Verified: true}, f.tenantA},
		"empty subject":      {security.Identity{Subject: "", Issuer: testIssuer, Audience: testAudience, Verified: true}, f.tenantA},
		"padded subject":     {ident(" " + subjectA), f.tenantA},
		"tenant not a uuid":  {good, "not-a-uuid"},
		"unknown tenant":     {good, "00000000-0000-4000-8000-000000000000"},
		"empty tenant":       {good, ""},
		"tenant with spaces": {good, f.tenantA + " "},
		"tenant injection":   {good, "x'; drop table tenants; --"},
		"subject of B on A":  {ident(subjectB), f.tenantA},
	}
	for name, c := range bad {
		if n, err := db.CountEvidence(ctx, c.id, c.tenant); err == nil || n != 0 {
			t.Fatalf("%s must fail with no data, got n=%d err=%v", name, n, err)
		}
	}
	if n, err := db.CountEvidence(ctx, good, f.tenantA); err != nil || n != 1 {
		t.Fatalf("tables damaged after hostile input: n=%d err=%v", n, err)
	}
}

func TestTransactionLocalSubjectDoesNotSurviveCommitRollbackOrError(t *testing.T) {
	f := needDB(t)
	ctx := context.Background()
	pool := newPool(t, f, 1) // a single connection forces reuse
	probe := func(when string) {
		var got string
		if err := pool.QueryRow(ctx, "select coalesce(current_setting('app.subject',true),'')").Scan(&got); err != nil || got != "" {
			t.Fatalf("subject %q leaked %s (err=%v)", got, when, err)
		}
	}
	db, _ := NewTenantDB(pool, testIssuer, testAudience)
	if _, err := db.CountEvidence(ctx, ident(subjectA), f.tenantA); err != nil {
		t.Fatal(err)
	}
	probe("after a committed read")
	if _, err := db.CountEvidence(ctx, ident(subjectA), f.tenantB); err == nil {
		t.Fatal("expected denial")
	}
	probe("after a membership denial (rollback)")
	_ = rawTx(ctx, pool, subjectA, true, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "select 1/0")
		return err
	})
	probe("after a SQL error")
	err := db.WithTenant(ctx, ident(subjectA), f.tenantA, func(pgx.Tx) error { return errors.New("callback failed") })
	if err == nil {
		t.Fatal("expected callback error")
	}
	probe("after a failing callback")
	// A callback that panics must not leave a subject on the pooled connection either.
	func() {
		defer func() { _ = recover() }()
		_ = db.WithTenant(ctx, ident(subjectA), f.tenantA, func(pgx.Tx) error { panic("boom") })
	}()
	probe("after a panicking callback")
}

func TestConcurrentPooledAccessNeverLeaksOrCrossesTenants(t *testing.T) {
	f := needDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	const maxConns = 4
	pool := newPool(t, f, maxConns) // far fewer connections than goroutines: constant reuse
	db, _ := NewTenantDB(pool, testIssuer, testAudience)

	var calls int64
	var wg sync.WaitGroup
	errs := make(chan error, 64)
	fail := func(format string, a ...any) {
		select {
		case errs <- fmt.Errorf(format, a...):
		default:
		}
	}
	for g := 0; g < 24; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				atomic.AddInt64(&calls, 1)
				switch (g + i) % 6 {
				case 0:
					if n, err := db.CountEvidence(ctx, ident(subjectA), f.tenantA); err != nil || n != 1 {
						fail("A on A: n=%d err=%v", n, err)
					}
				case 1:
					if n, err := db.CountEvidence(ctx, ident(subjectB), f.tenantB); err != nil || n != 1 {
						fail("B on B: n=%d err=%v", n, err)
					}
				case 2:
					if _, err := db.CountEvidence(ctx, ident(subjectA), f.tenantB); err == nil || err.Error() != "tenant membership denied" {
						fail("A on B must be denied, got %v", err)
					}
				case 3:
					if _, err := db.CountEvidence(ctx, ident(subjectB), f.tenantA); err == nil || err.Error() != "tenant membership denied" {
						fail("B on A must be denied, got %v", err)
					}
				case 4:
					// A bare pool read while other goroutines are inside tenant transactions.
					var subj string
					var n int
					if err := pool.QueryRow(ctx, "select coalesce(current_setting('app.subject',true),''), (select count(*) from evidence_metadata)").Scan(&subj, &n); err != nil || subj != "" || n != 0 {
						fail("no-subject probe saw subject=%q rows=%d err=%v", subj, n, err)
					}
				case 5:
					if _, err := db.CountEvidence(ctx, ident("nobody"), f.tenantA); err == nil || err.Error() != "tenant membership denied" {
						fail("unknown subject must be denied, got %v", err)
					}
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if t.Failed() {
		return
	}
	if atomic.LoadInt64(&calls) != 24*30 {
		t.Fatalf("expected %d calls, ran %d", 24*30, calls)
	}
	// Every pooled connection must be clean.
	conns := make([]*pgxpool.Conn, 0, maxConns)
	for i := 0; i < maxConns; i++ {
		c, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
		var subj string
		if err := c.QueryRow(ctx, "select coalesce(current_setting('app.subject',true),'')").Scan(&subj); err != nil || subj != "" {
			t.Fatalf("connection %d still carries subject %q (err=%v)", i, subj, err)
		}
	}
	for _, c := range conns {
		c.Release()
	}
}
