package eligibility

import "testing"

// s.46B (Act 862, in force 1 Jan 2025) requires a citizen and resident, a first residential property
// occupied as residence, an SPA dated 1 Jan 2025 to 31 Dec 2027, and no income from the property.
// Draft arithmetic and fail-closed checks only.
func TestFirstHomeEachRequiredFactBlocks(t *testing.T) {
	for _, rule := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		for _, fact := range ruleByID(t, rule).RequiredFacts {
			c := claim("a", "loan-1", rule, 300000)
			c.Facts[fact] = false
			r, err := Evaluate(Draft2025(), []Claim{c})
			if err != nil || r.ProvisionalTotalSen != 0 || r.Decisions[0].State != "REVIEW_REQUIRED" {
				t.Fatalf("%s/%s must block: %+v %v", rule, fact, r, err)
			}
		}
		need := map[string]bool{}
		for _, f := range ruleByID(t, rule).RequiredFacts {
			need[f] = true
		}
		for _, f := range []string{"citizen", "no_income_from_property", "occupied_as_residence", "qualifying_spa_date", "co_claimant_none", "within_three_years_of_spa"} {
			if !need[f] {
				t.Fatalf("%s lost required fact %s", rule, f)
			}
		}
	}
}

// s.46B also limits the relief to three consecutive basis years from the SPA date and
// apportions it where more than one person claims interest on the same property. Until
// the A x B / C apportionment is implemented, the claimant must confirm there is no
// co-claimant and that the claim year is within three years of the SPA. Missing or
// false facts fail closed; they never produce a provisional amount.
func TestFirstHomeCoClaimantAndThreeYearFactsFailClosed(t *testing.T) {
	for _, rule := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		for _, fact := range []string{"co_claimant_none", "within_three_years_of_spa"} {
			for _, missing := range []bool{true, false} {
				c := claim("a", "loan-1", rule, 300000)
				if missing {
					delete(c.Facts, fact)
				} else {
					c.Facts[fact] = false
				}
				r, err := Evaluate(Draft2025(), []Claim{c})
				if err != nil {
					t.Fatalf("%s/%s missing=%v: %v", rule, fact, missing, err)
				}
				d := r.Decisions[0]
				found := false
				for _, reason := range d.Reasons {
					if reason == "FACT_NOT_VERIFIED:"+fact {
						found = true
					}
				}
				if d.State != "REVIEW_REQUIRED" || r.ProvisionalTotalSen != 0 || r.ApprovedTotalSen != 0 || !found {
					t.Fatalf("%s/%s missing=%v must fail closed: %+v %+v", rule, fact, missing, r, d)
				}
			}
		}
	}
}

func TestFirstHomeCapBoundariesAndTiersExclusive(t *testing.T) {
	cases := []struct {
		rule         string
		amount, want int64
	}{
		{"FIRST_HOME_LOW", 699999, 699999}, {"FIRST_HOME_LOW", 700000, 700000}, {"FIRST_HOME_LOW", 700001, 700000},
		{"FIRST_HOME_MID", 499999, 499999}, {"FIRST_HOME_MID", 500000, 500000}, {"FIRST_HOME_MID", 500001, 500000},
	}
	for _, tc := range cases {
		r, err := Evaluate(Draft2025(), []Claim{claim("a", "loan-1", tc.rule, tc.amount)})
		if err != nil || r.ProvisionalTotalSen != tc.want || r.ApprovedTotalSen != 0 || r.FilingReady {
			t.Fatalf("%s %d: got %d err=%v", tc.rule, tc.amount, r.ProvisionalTotalSen, err)
		}
	}
	r, err := Evaluate(Draft2025(), []Claim{claim("a", "loan-1", "FIRST_HOME_LOW", 700000), claim("b", "loan-2", "FIRST_HOME_MID", 500000)})
	if err != nil || r.ProvisionalTotalSen != 700000 || r.Decisions[1].CandidateSen != 0 {
		t.Fatalf("tiers must be mutually exclusive: %+v %v", r, err)
	}
}
