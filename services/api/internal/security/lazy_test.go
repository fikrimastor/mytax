package security

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "sync"
 "sync/atomic"
 "testing"
 "time"
 "github.com/labstack/echo/v4"
)

type okVerifier struct{}

func (okVerifier) Verify(context.Context, string) (Identity, error) {
 return Identity{Subject: "lazy-user", Issuer: "https://example.test", Audience: "mytax-api", Verified: true}, nil
}

func newLazyForTest(t *testing.T, factory func(context.Context) (TokenVerifier, error), clock *time.Time) *LazyVerifier {
 t.Helper()
 l, err := NewLazyVerifier(factory, time.Second, 5*time.Second)
 if err != nil { t.Fatal(err) }
 l.now = func() time.Time { return *clock }
 return l
}

func TestLazyVerifierRetriesWithBackoffThenCaches(t *testing.T) {
 clock := time.Unix(1_000, 0)
 var calls int32
 up := false
 l := newLazyForTest(t, func(context.Context) (TokenVerifier, error) {
  atomic.AddInt32(&calls, 1)
  if !up { return nil, errors.New("keycloak not ready") }
  return okVerifier{}, nil
 }, &clock)

 // provider down: fail closed with ErrVerifierUnavailable, never success
 if id, err := l.Verify(context.Background(), "tok"); !errors.Is(err, ErrVerifierUnavailable) || id.Verified {
  t.Fatalf("want unavailable, got %+v %v", id, err)
 }
 // inside the backoff window the factory must not be hit again
 for i := 0; i < 5; i++ { l.Verify(context.Background(), "tok") }
 if n := atomic.LoadInt32(&calls); n != 1 { t.Fatalf("factory called %d times inside backoff, want 1", n) }

 // after the backoff, still down: exactly one more attempt
 clock = clock.Add(6 * time.Second)
 l.Verify(context.Background(), "tok")
 if n := atomic.LoadInt32(&calls); n != 2 { t.Fatalf("want 2 attempts, got %d", n) }

 // provider comes up; next window succeeds, then it is cached
 up = true
 clock = clock.Add(6 * time.Second)
 id, err := l.Verify(context.Background(), "tok")
 if err != nil || !id.Verified { t.Fatalf("expected success after recovery: %+v %v", id, err) }
 for i := 0; i < 5; i++ { l.Verify(context.Background(), "tok") }
 if n := atomic.LoadInt32(&calls); n != 3 { t.Fatalf("factory must not be called after success, calls=%d", n) }
}

func TestLazyVerifierSingleFlightUnderConcurrency(t *testing.T) {
 clock := time.Unix(1_000, 0)
 var calls int32
 release := make(chan struct{})
 started := make(chan struct{})
 l := newLazyForTest(t, func(context.Context) (TokenVerifier, error) {
  if atomic.AddInt32(&calls, 1) == 1 { close(started) }
  <-release
  return nil, errors.New("still down")
 }, &clock)

 var wg sync.WaitGroup
 first := make(chan error, 1)
 wg.Add(1)
 go func() { defer wg.Done(); _, err := l.Verify(context.Background(), "tok"); first <- err }()
 <-started // one attempt is now in flight
 var unavailable int32
 for i := 0; i < 20; i++ {
  wg.Add(1)
  go func() { defer wg.Done(); if _, err := l.Verify(context.Background(), "tok"); errors.Is(err, ErrVerifierUnavailable) { atomic.AddInt32(&unavailable, 1) } }()
 }
 // the 20 others return immediately (fast 503) while the attempt is blocked
 deadline := time.Now().Add(2 * time.Second)
 for atomic.LoadInt32(&unavailable) < 20 && time.Now().Before(deadline) { time.Sleep(5 * time.Millisecond) }
 if got := atomic.LoadInt32(&unavailable); got != 20 { t.Fatalf("concurrent callers should fail fast, %d/20 did", got) }
 close(release)
 wg.Wait()
 if err := <-first; !errors.Is(err, ErrVerifierUnavailable) { t.Fatalf("first caller: %v", err) }
 if n := atomic.LoadInt32(&calls); n != 1 { t.Fatalf("factory ran %d times concurrently, want 1", n) }
}

func TestLazyVerifierFactoryTimeoutAndNil(t *testing.T) {
 clock := time.Unix(1_000, 0)
 l := newLazyForTest(t, func(ctx context.Context) (TokenVerifier, error) {
  <-ctx.Done() // factory honours the attempt timeout
  return nil, ctx.Err()
 }, &clock)
 start := time.Now()
 if _, err := l.Verify(context.Background(), "tok"); !errors.Is(err, ErrVerifierUnavailable) { t.Fatalf("got %v", err) }
 if time.Since(start) > 3*time.Second { t.Fatal("attempt timeout not applied") }

 clock2 := time.Unix(1_000, 0)
 l2 := newLazyForTest(t, func(context.Context) (TokenVerifier, error) { return nil, nil }, &clock2)
 if _, err := l2.Verify(context.Background(), "tok"); !errors.Is(err, ErrVerifierUnavailable) { t.Fatalf("nil verifier must be unavailable, got %v", err) }

 if _, err := NewLazyVerifier(nil, time.Second, time.Second); err == nil { t.Fatal("nil factory accepted") }
 if _, err := NewLazyVerifier(func(context.Context) (TokenVerifier, error) { return okVerifier{}, nil }, 0, time.Second); err == nil { t.Fatal("zero timeout accepted") }
}

func TestRequireAccessReturns503WhenVerifierUnavailable(t *testing.T) {
 clock := time.Unix(1_000, 0)
 l := newLazyForTest(t, func(context.Context) (TokenVerifier, error) { return nil, errors.New("down") }, &clock)
 e := echo.New()
 e.GET("/p", RequireAccess(l, func(c echo.Context) error { return c.NoContent(200) }))
 req := httptest.NewRequest(http.MethodGet, "/p", nil)
 req.Header.Set("Authorization", "Bearer anything")
 rec := httptest.NewRecorder()
 e.ServeHTTP(rec, req)
 if rec.Code != http.StatusServiceUnavailable { t.Fatalf("got %d want 503", rec.Code) }
 // a missing token is still 401 regardless of provider state
 rec = httptest.NewRecorder()
 e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/p", nil))
 if rec.Code != http.StatusUnauthorized { t.Fatalf("no token: got %d want 401", rec.Code) }
}

func TestValidateAccessConfigFailsFast(t *testing.T) {
 if err := ValidateAccessConfig("https://idp.example/realms/x", "aud", "client", "scope"); err != nil { t.Fatal(err) }
 for _, c := range [][4]string{{"http://idp", "a", "c", "s"}, {"", "a", "c", "s"}, {"https://i", "", "c", "s"}, {"https://i", "a", "", "s"}, {"https://i", "a", "c", ""}} {
  if ValidateAccessConfig(c[0], c[1], c[2], c[3]) == nil { t.Fatalf("accepted bad config %v", c) }
 }
}
