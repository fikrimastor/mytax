package policy

import "testing"

// Status Approved in this file is an arithmetic fixture, not tax-policy sign-off.

func fixedPackage() Package {
	return Package{AssessmentYear: 2025, Version: "t", Status: Approved,
		Groups: []Group{{ID: "G", LimitSen: 1000}},
		Rules: []Rule{
			{ID: "FIX", FixedAmount: true, LimitSen: 700000, Sources: testSources},
			{ID: "EXP", Group: "G", LimitSen: 1000, Sources: testSources},
		}}
}

func TestFixedAmountClaimReceivesTheRuleLimit(t *testing.T) {
	r, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", 0)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000}, 700000)
	if hasReason(r.Decisions[0], "LIMIT_REACHED") {
		t.Fatalf("fixed claim must not be LIMIT_REACHED: %+v", r.Decisions[0])
	}
}

func TestFixedAmountClaimRejectsAnySentAmount(t *testing.T) {
	for _, amt := range []int64{1, -1, 700000} {
		if _, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", amt)}); err == nil {
			t.Fatalf("amount %d accepted on a fixed-amount rule", amt)
		}
	}
}

func TestExpenseRuleStillRejectsZeroAmount(t *testing.T) {
	if _, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "EXP", 0)}); err == nil {
		t.Fatal("zero amount accepted on an expense rule")
	}
}

func TestSecondFixedAmountClaimGetsZero(t *testing.T) {
	r, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", 0), synthClaim("b", "FIX", 0)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000, 0}, 700000)
	if !hasReason(r.Decisions[1], "LIMIT_REACHED") {
		t.Fatalf("second fixed claim: %+v", r.Decisions[1])
	}
}

func TestFixedAmountRuleStructureIsValidated(t *testing.T) {
	if err := ValidateRelease(fixedPackage(), testRecord(fixedPackage())); err != nil {
		t.Fatalf("valid fixed package rejected: %v", err)
	}
	cases := map[string]func(*Package){
		"zero limit": func(p *Package) { p.Rules[0].LimitSen = 0 },
		"in a group": func(p *Package) { p.Rules[0].Group = "G" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := fixedPackage()
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
