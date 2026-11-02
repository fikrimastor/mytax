package policy

import "testing"

func TestFirstHomeEachRequiredFactBlocks(t *testing.T) {
	p := YA2025Draft()
	for _, id := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		for _, fact := range findRule(t, p, id).RequiredFacts {
			c := claimFor(t, p, "a", id, 300000)
			c.Facts[fact] = false
			r, err := Evaluate(p, []Claim{c})
			if err != nil || r.ProvisionalTotalSen != 0 || r.Decisions[0].Status != "REVIEW_REQUIRED" {
				t.Fatalf("%s/%s must block: %+v %v", id, fact, r, err)
			}
		}
	}
}

func TestFirstHomeCapBoundaries(t *testing.T) {
	p := YA2025Draft()
	cases := []struct {
		rule         string
		amount, want int64
	}{
		{"FIRST_HOME_LOW", 699999, 699999}, {"FIRST_HOME_LOW", 700000, 700000}, {"FIRST_HOME_LOW", 700001, 700000},
		{"FIRST_HOME_MID", 499999, 499999}, {"FIRST_HOME_MID", 500000, 500000}, {"FIRST_HOME_MID", 500001, 500000},
	}
	for _, tc := range cases {
		r, err := Evaluate(p, []Claim{claimFor(t, p, "a", tc.rule, tc.amount)})
		if err != nil || r.ProvisionalTotalSen != tc.want || r.ApprovedTotalSen != 0 || r.FilingReady {
			t.Fatalf("%s %d: got %d err=%v", tc.rule, tc.amount, r.ProvisionalTotalSen, err)
		}
	}
}
