package policy

import (
	"fmt"
	"math"
	"testing"
)

func TestExclusiveFamilyBlocksTheSecondFirstHomeTier(t *testing.T) {
	p := YA2025Draft()
	low := claimFor(t, p, "a", "FIRST_HOME_LOW", 300000)
	mid := claimFor(t, p, "b", "FIRST_HOME_MID", 300000)
	r, err := Evaluate(p, []Claim{low, mid})
	if err != nil || r.ProvisionalTotalSen != 300000 {
		t.Fatalf("%+v %v", r, err)
	}
	if r.Decisions[1].ProvisionalAmountSen != 0 || r.Decisions[1].Reasons[0] != "MUTUALLY_EXCLUSIVE_RELIEF" {
		t.Fatalf("second tier must be blocked: %+v", r.Decisions[1])
	}
}

func TestExpenseKeyIsRequiredAndCannotBeReused(t *testing.T) {
	p := YA2025Draft()
	none := claimFor(t, p, "a", "LIFESTYLE_BOOKS", 100)
	none.ExpenseKey = ""
	r, _ := Evaluate(p, []Claim{none})
	if r.Decisions[0].Reasons[0] != "EXPENSE_IDENTIFIER_MISSING" || r.ProvisionalTotalSen != 0 {
		t.Fatalf("missing key must block: %+v", r.Decisions[0])
	}
	a := claimFor(t, p, "a", "LIFESTYLE_BOOKS", 100)
	b := claimFor(t, p, "b", "LIFESTYLE_INTERNET", 100)
	b.ExpenseKey = a.ExpenseKey
	r, _ = Evaluate(p, []Claim{a, b})
	if r.ProvisionalTotalSen != 100 || r.Decisions[1].Reasons[0] != "DUPLICATE_EXPENSE" {
		t.Fatalf("reused key must block the second claim: %+v", r)
	}
}

func TestClaimCountLimit(t *testing.T) {
	p := YA2025Draft()
	mk := func(n int) []Claim {
		out := make([]Claim, 0, n)
		for i := 0; i < n; i++ {
			c := claimFor(t, p, fmt.Sprintf("c%04d", i), "LIFESTYLE_BOOKS", 1)
			out = append(out, c)
		}
		return out
	}
	if _, err := Evaluate(p, mk(1000)); err != nil {
		t.Fatalf("1000 claims must be accepted: %v", err)
	}
	if _, err := Evaluate(p, mk(1001)); err == nil {
		t.Fatal("1001 claims accepted")
	}
}

func TestOnlyYear2025IsEvaluated(t *testing.T) {
	for _, year := range []int{2023, 2024, 2026} {
		p := YA2025Draft()
		p.AssessmentYear = year
		if _, err := Evaluate(p, nil); err == nil {
			t.Fatalf("year %d accepted", year)
		}
	}
}

func TestRunningTotalOverflowIsRejected(t *testing.T) {
	// Real catalog limits cannot reach int64 overflow, so use a synthetic package.
	p := Package{AssessmentYear: 2025, Version: "synthetic", Status: Draft, Rules: []Rule{
		{ID: "R1", LimitSen: math.MaxInt64}, {ID: "R2", LimitSen: math.MaxInt64},
	}}
	claims := []Claim{
		{ID: "a", RuleID: "R1", AmountSen: math.MaxInt64, ExpenseKey: "k1", EvidenceReference: "x"},
		{ID: "b", RuleID: "R2", AmountSen: math.MaxInt64, ExpenseKey: "k2", EvidenceReference: "x"},
	}
	if _, err := Evaluate(p, claims); err == nil {
		t.Fatal("overflowing total accepted")
	}
}

func TestDraftCatalogNeverProducesApprovedAmounts(t *testing.T) {
	p := YA2025Draft()
	for i, rule := range p.Rules {
		c := claimFor(t, p, "a", rule.ID, 1)
		c.ExpenseKey = fmt.Sprintf("k%d", i)
		r, err := Evaluate(p, []Claim{c})
		if err != nil {
			t.Fatalf("%s: %v", rule.ID, err)
		}
		if r.ApprovedTotalSen != 0 || r.Decisions[0].ApprovedAmountSen != 0 || r.FilingReady || r.Decisions[0].Status != "REVIEW_REQUIRED" {
			t.Fatalf("%s leaked approval: %+v", rule.ID, r)
		}
	}
}
