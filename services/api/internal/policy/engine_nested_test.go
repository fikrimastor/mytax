package policy

import "testing"

// Status Approved in this file is an arithmetic fixture, not tax-policy sign-off.

var testSources = []Source{{URL: "https://www.hasil.gov.my/test", Section: "synthetic", RetrievedOn: "2026-11-16"}}

func testRecord(p Package) ReleaseRecord {
	return ReleaseRecord{PackageVersion: p.Version, AssessmentYear: p.AssessmentYear, ReviewerID: "r", ReviewedAt: "2026-11-16",
		ApprovalReference: "synthetic", SourceSnapshotSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
}

func nestedPackage() Package {
	return Package{AssessmentYear: 2025, Version: "t", Status: Approved,
		Groups: []Group{{ID: "TOP", LimitSen: 1000}, {ID: "SUB", Parent: "TOP", LimitSen: 300}},
		Rules: []Rule{
			{ID: "BIG", Group: "TOP", LimitSen: 1000, Sources: testSources},
			{ID: "S1", Group: "SUB", LimitSen: 300, Sources: testSources},
			{ID: "S2", Group: "SUB", LimitSen: 300, Sources: testSources},
		}}
}

func synthClaim(id, rule string, amt int64) Claim {
	return Claim{ID: id, RuleID: rule, AmountSen: amt, ExpenseKey: "exp-" + id, EvidenceReference: "x"}
}

func wantAmounts(t *testing.T, r Result, want []int64, total int64) {
	t.Helper()
	if len(r.Decisions) != len(want) {
		t.Fatalf("%d decisions, want %d", len(r.Decisions), len(want))
	}
	for i, w := range want {
		if got := r.Decisions[i].ProvisionalAmountSen; got != w {
			t.Fatalf("claim %s = %d, want %d (%v)", r.Decisions[i].ClaimID, got, w, r.Decisions[i].Reasons)
		}
	}
	if r.ProvisionalTotalSen != total {
		t.Fatalf("total %d, want %d", r.ProvisionalTotalSen, total)
	}
}

func hasReason(d Decision, reason string) bool {
	for _, r := range d.Reasons {
		if r == reason {
			return true
		}
	}
	return false
}

func TestChildGroupCapsItsRulesAndParentKeepsTheRest(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "S1", 200), synthClaim("b", "S2", 200), synthClaim("c", "BIG", 900)})
	if err != nil {
		t.Fatal(err)
	}
	// SUB is full after 300; TOP has 700 left for BIG.
	wantAmounts(t, r, []int64{200, 100, 700}, 1000)
	if !hasReason(r.Decisions[1], "LIMIT_REACHED") || !hasReason(r.Decisions[2], "LIMIT_REACHED") {
		t.Fatalf("missing LIMIT_REACHED: %+v", r.Decisions)
	}
}

func TestFullChildGroupGivesZeroWhileParentStillPays(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "S1", 300), synthClaim("b", "S2", 50), synthClaim("c", "BIG", 1000)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{300, 0, 700}, 1000)
}

func TestParentGroupCapBindsChildRules(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "BIG", 900), synthClaim("b", "S1", 300)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{900, 100}, 1000)
}

func TestNestedGroupStructureIsValidated(t *testing.T) {
	if err := ValidateRelease(nestedPackage(), testRecord(nestedPackage())); err != nil {
		t.Fatalf("valid nested package rejected: %v", err)
	}
	cases := map[string]func(*Package){
		"unknown parent":           func(p *Package) { p.Groups[1].Parent = "NOPE" },
		"self parent":              func(p *Package) { p.Groups[0].Parent = "TOP" },
		"two-group cycle":          func(p *Package) { p.Groups[0].Parent = "SUB"; p.Groups[1].LimitSen = 1000 },
		"child limit above parent": func(p *Package) { p.Groups[1].LimitSen = 1001 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := nestedPackage()
			mutate(&p)
			if _, err := Evaluate(p, nil); err == nil {
				t.Fatal("Evaluate accepted the package")
			}
			if err := ValidateRelease(p, testRecord(p)); err == nil {
				t.Fatal("ValidateRelease accepted the package")
			}
		})
	}
}
