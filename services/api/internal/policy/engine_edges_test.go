package policy

import "testing"

func TestBlockedClaimDoesNotSetExclusiveFamily(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "FIRST_HOME_LOW", 300000)
	a.Facts[findRule(t, p, "FIRST_HOME_LOW").RequiredFacts[0]] = false // blocked by an unverified fact
	b := claimFor(t, p, "b", "FIRST_HOME_MID", 300000)
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decisions[0].ProvisionalAmountSen != 0 {
		t.Fatalf("blocked claim a got amount %d", r.Decisions[0].ProvisionalAmountSen)
	}
	if r.Decisions[1].ProvisionalAmountSen != 300000 || r.ProvisionalTotalSen != 300000 {
		t.Fatalf("b=%d total=%d, want 300000 each: %+v", r.Decisions[1].ProvisionalAmountSen, r.ProvisionalTotalSen, r.Decisions[1])
	}
}

func TestSetupValidationErrors(t *testing.T) {
	base := func() Package {
		return Package{AssessmentYear: 2025, Version: "t", Status: Draft,
			Groups: []Group{{ID: "G", LimitSen: 100}},
			Rules:  []Rule{{ID: "R", Group: "G", LimitSen: 100}}}
	}
	cases := map[string]func(*Package){
		"negative group limit":   func(p *Package) { p.Groups[0].LimitSen = -1 },
		"duplicate group id":     func(p *Package) { p.Groups = append(p.Groups, Group{ID: "G", LimitSen: 1}) },
		"empty group id":         func(p *Package) { p.Groups[0].ID = "" },
		"negative rule limit":    func(p *Package) { p.Rules[0].LimitSen = -1 },
		"empty rule id":          func(p *Package) { p.Rules[0].ID = "" },
		"duplicate rule id":      func(p *Package) { p.Rules = append(p.Rules, Rule{ID: "R", LimitSen: 1}) },
		"unknown group":          func(p *Package) { p.Rules[0].Group = "NOPE" },
		"empty package version":  func(p *Package) { p.Version = "" },
		"unknown package status": func(p *Package) { p.Status = "X" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := base()
			mutate(&p)
			if _, err := Evaluate(p, nil); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
	t.Run("valid minimal package", func(t *testing.T) {
		if _, err := Evaluate(base(), nil); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRuleLimitAndGroupLimitBothApply(t *testing.T) {
	// Status Approved here is an arithmetic fixture, not tax-policy sign-off.
	p := Package{AssessmentYear: 2025, Version: "t", Status: Approved,
		Groups: []Group{{ID: "G", LimitSen: 1000}},
		Rules: []Rule{
			{ID: "R1", Group: "G", LimitSen: 400},
			{ID: "R2", Group: "G", LimitSen: 900},
		}}
	mk := func(id, rule string, amt int64) Claim {
		return Claim{ID: id, RuleID: rule, AmountSen: amt, ExpenseKey: "exp-" + id, EvidenceReference: "x"}
	}
	r, err := Evaluate(p, []Claim{mk("c", "R2", 100), mk("a", "R1", 700), mk("b", "R2", 900)})
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{400, 600, 0}
	for i, w := range want {
		d := r.Decisions[i]
		if d.ProvisionalAmountSen != w || d.ApprovedAmountSen != w {
			t.Fatalf("claim %s provisional=%d approved=%d, want %d", d.ClaimID, d.ProvisionalAmountSen, d.ApprovedAmountSen, w)
		}
	}
	if r.ApprovedTotalSen != 1000 {
		t.Fatalf("approved total %d, want 1000", r.ApprovedTotalSen)
	}
	hasLimit := false
	for _, reason := range r.Decisions[2].Reasons {
		if reason == "LIMIT_REACHED" {
			hasLimit = true
		}
	}
	if !hasLimit {
		t.Fatalf("claim c missing LIMIT_REACHED: %+v", r.Decisions[2])
	}
}

// Relief is once per family, same as the retired eligibility engine, so a
// second claim in the same first-home tier is blocked. Fails closed.
func TestSecondClaimInSameFirstHomeTierIsBlocked(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "FIRST_HOME_LOW", 300000)
	b := claimFor(t, p, "b", "FIRST_HOME_LOW", 300000)
	b.ExpenseKey = "exp-b-other"
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decisions[0].ProvisionalAmountSen != 300000 {
		t.Fatalf("a = %+v", r.Decisions[0])
	}
	if r.Decisions[1].ProvisionalAmountSen != 0 || len(r.Decisions[1].Reasons) == 0 || r.Decisions[1].Reasons[0] != "MUTUALLY_EXCLUSIVE_RELIEF" {
		t.Fatalf("b = %+v", r.Decisions[1])
	}
	if r.ProvisionalTotalSen != 300000 {
		t.Fatalf("total = %d", r.ProvisionalTotalSen)
	}
}
