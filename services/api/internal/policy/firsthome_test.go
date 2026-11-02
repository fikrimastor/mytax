package policy

import "testing"

// s.46B (Act 53, as replaced by Finance Act 2024 s.10) limits first-home interest
// relief to three consecutive basis years from the SPA date and shares it between
// co-claimants by the A x B / C formula. Apportionment is not implemented, so the
// draft catalog requires an explicit no-co-claimant confirmation and a claim year
// within three years of the SPA. Both are fail-closed facts.
func TestFirstHomeRequiresCoClaimantAndThreeYearFacts(t *testing.T) {
	p := YA2025Draft()
	for _, id := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		rule := findRule(t, p, id)
		need := map[string]bool{}
		for _, f := range rule.RequiredFacts {
			need[f] = true
		}
		for _, fact := range []string{"co_claimant_none", "within_three_years_of_spa"} {
			if !need[fact] {
				t.Fatalf("%s lost required fact %s: %v", id, fact, rule.RequiredFacts)
			}
		}

		// Positive control: with every required fact present the draft still only
		// produces a provisional amount.
		ok := claimFor(t, p, "a", id, 100000)
		r, err := Evaluate(p, []Claim{ok})
		if err != nil || r.Decisions[0].ProvisionalAmountSen != 100000 || r.ApprovedTotalSen != 0 {
			t.Fatalf("%s: complete facts should be provisional only: %+v %v", id, r, err)
		}

		for _, fact := range []string{"co_claimant_none", "within_three_years_of_spa"} {
			for _, missing := range []bool{true, false} {
				c := claimFor(t, p, "a", id, 100000)
				if missing {
					delete(c.Facts, fact)
				} else {
					c.Facts[fact] = false
				}
				res, err := Evaluate(p, []Claim{c})
				if err != nil {
					t.Fatalf("%s/%s: %v", id, fact, err)
				}
				d := res.Decisions[0]
				found := false
				for _, reason := range d.Reasons {
					if reason == "UNVERIFIED_FACT:"+fact {
						found = true
					}
				}
				if d.Status != "REVIEW_REQUIRED" || d.ProvisionalAmountSen != 0 || d.ApprovedAmountSen != 0 || res.ApprovedTotalSen != 0 || !found {
					t.Fatalf("%s/%s missing=%v must fail closed: %+v", id, fact, missing, d)
				}
			}
		}
	}
}
