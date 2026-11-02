package policy

import "testing"

// Lifestyle (YA2025 item 9) tests for the policy package. The Approved status used
// in some tests is an arithmetic fixture only; it is NOT tax-policy sign-off.

const lifestyleCapSen = int64(250000)

var lifestyleRuleIDs = []string{"LIFESTYLE_BOOKS", "LIFESTYLE_DEVICE", "LIFESTYLE_INTERNET", "LIFESTYLE_COURSE"}

func findRule(t *testing.T, p Package, id string) Rule {
	t.Helper()
	for _, r := range p.Rules {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("rule %s missing", id)
	return Rule{}
}

func claimFor(t *testing.T, p Package, id, rule string, amt int64) Claim {
	t.Helper()
	f := map[string]bool{}
	for _, k := range findRule(t, p, rule).RequiredFacts {
		f[k] = true
	}
	return Claim{ID: id, RuleID: rule, AmountSen: amt, EvidencePresent: true, Facts: f}
}

func TestLifestyleRulesHaveGroupCapAndSources(t *testing.T) {
	p := YA2025Draft()
	for _, id := range lifestyleRuleIDs {
		r := findRule(t, p, id)
		if r.Group != "LIFESTYLE" || r.LimitSen != lifestyleCapSen {
			t.Fatalf("%s group=%s limit=%d", id, r.Group, r.LimitSen)
		}
		if len(r.Sources) == 0 {
			t.Fatalf("%s has no source", id)
		}
		for _, s := range r.Sources {
			if s.URL == "" || s.Section == "" || s.RetrievedOn == "" {
				t.Fatalf("%s incomplete source %+v", id, s)
			}
		}
		if len(r.RequiredFacts) < 4 {
			t.Fatalf("%s has too few required facts: %v", id, r.RequiredFacts)
		}
	}
	for _, g := range p.Groups {
		if g.ID == "LIFESTYLE" && g.LimitSen != lifestyleCapSen {
			t.Fatalf("group cap %d", g.LimitSen)
		}
	}
}

func TestLifestyleDraftNeverApproves(t *testing.T) {
	p := YA2025Draft()
	for _, id := range lifestyleRuleIDs {
		r, err := Evaluate(p, []Claim{claimFor(t, p, "a", id, 100000)})
		if err != nil {
			t.Fatal(err)
		}
		d := r.Decisions[0]
		if r.ApprovedTotalSen != 0 || d.ApprovedAmountSen != 0 || d.Status != "REVIEW_REQUIRED" || d.ProvisionalAmountSen != 100000 {
			t.Fatalf("%s draft leaked approval: %+v", id, r)
		}
	}
}

func TestLifestyleCapBoundaries(t *testing.T) {
	p := YA2025Draft()
	p.Status = Approved // arithmetic fixture only
	cases := []struct{ amount, want int64 }{
		{1, 1}, {lifestyleCapSen - 1, lifestyleCapSen - 1}, {lifestyleCapSen, lifestyleCapSen},
		{lifestyleCapSen + 1, lifestyleCapSen}, {lifestyleCapSen * 10, lifestyleCapSen},
	}
	for _, id := range lifestyleRuleIDs {
		for _, tc := range cases {
			r, err := Evaluate(p, []Claim{claimFor(t, p, "a", id, tc.amount)})
			if err != nil || r.ApprovedTotalSen != tc.want || r.Decisions[0].ProvisionalAmountSen != tc.want {
				t.Fatalf("%s amount=%d got=%d want=%d err=%v", id, tc.amount, r.ApprovedTotalSen, tc.want, err)
			}
		}
	}
}

func TestLifestyleSharedCapAcrossFourCategoriesAndOrder(t *testing.T) {
	p := YA2025Draft()
	p.Status = Approved // arithmetic fixture only
	claims := []Claim{
		claimFor(t, p, "d", "LIFESTYLE_COURSE", 70000),
		claimFor(t, p, "b", "LIFESTYLE_INTERNET", 90000),
		claimFor(t, p, "c", "LIFESTYLE_DEVICE", 80000),
		claimFor(t, p, "a", "LIFESTYLE_BOOKS", 60000),
	}
	r, err := Evaluate(p, claims)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{"a": 60000, "b": 90000, "c": 80000, "d": 20000}
	for _, d := range r.Decisions {
		if d.ApprovedAmountSen != want[d.ClaimID] {
			t.Fatalf("claim %s got %d want %d", d.ClaimID, d.ApprovedAmountSen, want[d.ClaimID])
		}
	}
	if r.ApprovedTotalSen != lifestyleCapSen {
		t.Fatalf("total %d", r.ApprovedTotalSen)
	}
	rev := []Claim{claims[3], claims[2], claims[1], claims[0]}
	r2, _ := Evaluate(p, rev)
	if r2.ApprovedTotalSen != r.ApprovedTotalSen {
		t.Fatal("input order changed total")
	}
	for _, d := range r2.Decisions {
		if d.ApprovedAmountSen != want[d.ClaimID] {
			t.Fatalf("order changed claim %s: %d", d.ClaimID, d.ApprovedAmountSen)
		}
	}
}

func TestLifestyleEachRequiredFactBlocksAndDoesNotConsumeCap(t *testing.T) {
	p := YA2025Draft()
	p.Status = Approved // arithmetic fixture only
	for _, id := range lifestyleRuleIDs {
		for _, fact := range findRule(t, p, id).RequiredFacts {
			bad := claimFor(t, p, "a", id, 100000)
			bad.Facts[fact] = false
			good := claimFor(t, p, "b", "LIFESTYLE_BOOKS", lifestyleCapSen)
			r, err := Evaluate(p, []Claim{bad, good})
			if err != nil {
				t.Fatal(err)
			}
			d := r.Decisions[0]
			found := false
			for _, reason := range d.Reasons {
				if reason == "UNVERIFIED_FACT:"+fact {
					found = true
				}
			}
			if d.Status != "REVIEW_REQUIRED" || d.ApprovedAmountSen != 0 || d.ProvisionalAmountSen != 0 || !found {
				t.Fatalf("%s/%s must block: %+v", id, fact, d)
			}
			if r.Decisions[1].ApprovedAmountSen != lifestyleCapSen {
				t.Fatalf("%s/%s blocked claim consumed shared cap", id, fact)
			}
		}
	}
}

func TestLifestyleEvidenceDuplicateAndUnknownRuleFailClosed(t *testing.T) {
	p := YA2025Draft()
	p.Status = Approved // arithmetic fixture only
	for _, id := range lifestyleRuleIDs {
		c := claimFor(t, p, "a", id, 100000)
		c.EvidencePresent = false
		r, _ := Evaluate(p, []Claim{c})
		if r.ApprovedTotalSen != 0 || r.Decisions[0].Reasons[0] != "EVIDENCE_MISSING" {
			t.Fatalf("%s: missing evidence must block", id)
		}
		c = claimFor(t, p, "a", id, 100000)
		c.Duplicate = true
		r, _ = Evaluate(p, []Claim{c})
		if r.ApprovedTotalSen != 0 || r.Decisions[0].Reasons[0] != "DUPLICATE_EXPENSE" {
			t.Fatalf("%s: duplicate must block", id)
		}
	}
	r, _ := Evaluate(p, []Claim{{ID: "a", RuleID: "LIFESTYLE_UNKNOWN", AmountSen: 1, EvidencePresent: true}})
	if r.ApprovedTotalSen != 0 || r.Decisions[0].Reasons[0] != "UNKNOWN_RELIEF_RULE" {
		t.Fatalf("unknown rule must fail closed: %+v", r.Decisions[0])
	}
}

func TestLifestyleRejectsBadInputAndYear(t *testing.T) {
	p := YA2025Draft()
	for _, amt := range []int64{0, -1} {
		if _, err := Evaluate(p, []Claim{claimFor(t, p, "a", "LIFESTYLE_DEVICE", amt)}); err == nil {
			t.Fatalf("amount %d accepted", amt)
		}
	}
	if _, err := Evaluate(p, []Claim{claimFor(t, p, "a", "LIFESTYLE_DEVICE", 1), claimFor(t, p, "a", "LIFESTYLE_COURSE", 1)}); err == nil {
		t.Fatal("duplicate claim ID accepted")
	}
	for _, year := range []int{1999, 2101} {
		q := YA2025Draft()
		q.AssessmentYear = year
		if _, err := Evaluate(q, nil); err == nil {
			t.Fatalf("year %d accepted", year)
		}
	}
}
