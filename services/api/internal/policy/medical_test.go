package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Golden fixtures from LHDN Public Ruling No. 7/2025 Examples 3-20 (excerpt in docs/sources/pr7-2025).
// Examples 12 and 15 need joint assessment (s.46(3)) and are not covered.

type pr7Line struct {
	ID        string `json:"id"`
	AmountSen int64  `json:"amount_sen"`
	Reason    string `json:"reason"`
}

type pr7Case struct {
	Example  int             `json:"example"`
	Taxpayer string          `json:"taxpayer"`
	Claims   []recordedClaim `json:"claims"`
	Expect   struct {
		Lines    []pr7Line `json:"lines"`
		TotalSen int64     `json:"total_sen"`
	} `json:"expect"`
}

func TestPR72025WorkedExamples(t *testing.T) {
	raw, err := os.ReadFile("testdata/pr7_2025_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []pr7Case
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 17 {
		t.Fatalf("%d cases, want 17", len(cases))
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("example %d %s", c.Example, c.Taxpayer), func(t *testing.T) {
			p := YA2025Draft()
			r, err := Evaluate(p, buildRecordedClaims(p, recordedCase{Claims: c.Claims}))
			if err != nil {
				t.Fatal(err)
			}
			want := make([]int64, len(c.Expect.Lines))
			for i, l := range c.Expect.Lines {
				want[i] = l.AmountSen
			}
			wantAmounts(t, r, want, c.Expect.TotalSen)
			for i, l := range c.Expect.Lines {
				d := r.Decisions[i]
				if d.ClaimID != l.ID || d.Status != "REVIEW_REQUIRED" || d.ApprovedAmountSen != 0 {
					t.Fatalf("line %d: %+v", i, d)
				}
				if l.Reason != "" && !hasReason(d, l.Reason) {
					t.Fatalf("line %s lacks %s: %v", l.ID, l.Reason, d.Reasons)
				}
			}
		})
	}
}

func TestMedicalCapBoundaries(t *testing.T) {
	caps := map[string]int64{
		"PARENTS_MEDICAL": 800000, "PARENTS_CARER": 800000, "PARENTS_CHECKUP": 100000,
		"DISABILITY_EQUIPMENT": 600000, "MEDICAL_SERIOUS_DISEASE": 1000000, "MEDICAL_FERTILITY": 1000000,
		"MEDICAL_VACCINATION": 100000, "MEDICAL_DENTAL": 100000, "MEDICAL_LEARNING_DISABILITY": 600000,
		"MEDICAL_SCREENING_EXAM": 100000, "MEDICAL_SCREENING_DETECTION_TEST": 100000,
		"MEDICAL_SCREENING_SELF_TEST_DEVICE": 100000, "MEDICAL_SCREENING_MENTAL_HEALTH": 100000,
	}
	p := YA2025Draft()
	for rule, cap := range caps {
		t.Run(rule, func(t *testing.T) {
			at, err := Evaluate(p, []Claim{claimFor(t, p, "a", rule, cap)})
			if err != nil {
				t.Fatal(err)
			}
			wantAmounts(t, at, []int64{cap}, cap)
			if hasReason(at.Decisions[0], "LIMIT_REACHED") {
				t.Fatalf("at cap flagged: %v", at.Decisions[0].Reasons)
			}
			over, err := Evaluate(p, []Claim{claimFor(t, p, "a", rule, cap+1)})
			if err != nil {
				t.Fatal(err)
			}
			wantAmounts(t, over, []int64{cap}, cap)
			if !hasReason(over.Decisions[0], "LIMIT_REACHED") {
				t.Fatalf("one sen over not flagged: %v", over.Decisions[0].Reasons)
			}
		})
	}
}

func TestMedicalGroupBoundaries(t *testing.T) {
	cases := []struct {
		name         string
		first, other string
		firstAmt     int64
		rest         int64 // amount left in the group after the first claim
	}{
		{"PARENTS RM8,000", "PARENTS_MEDICAL", "PARENTS_CARER", 500000, 300000},
		{"MEDICAL_SELF_FAMILY RM10,000", "MEDICAL_SERIOUS_DISEASE", "MEDICAL_FERTILITY", 900000, 100000},
		{"MEDICAL_SCREENING RM1,000", "MEDICAL_SCREENING_EXAM", "MEDICAL_SCREENING_DETECTION_TEST", 60000, 40000},
	}
	p := YA2025Draft()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, extra := range []int64{0, 1} {
				r, err := Evaluate(p, []Claim{claimFor(t, p, "a", tc.first, tc.firstAmt), claimFor(t, p, "b", tc.other, tc.rest+extra)})
				if err != nil {
					t.Fatal(err)
				}
				wantAmounts(t, r, []int64{tc.firstAmt, tc.rest}, tc.firstAmt+tc.rest)
				if hasReason(r.Decisions[1], "LIMIT_REACHED") != (extra == 1) {
					t.Fatalf("extra %d: %v", extra, r.Decisions[1].Reasons)
				}
			}
		})
	}
}

func TestDisabledSelfIsAFixedRM7000(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "DISABLED_SELF", 0)
	b := claimFor(t, p, "b", "DISABLED_SELF", 0)
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000, 0}, 700000)
	if _, err := Evaluate(p, []Claim{claimFor(t, p, "a", "DISABLED_SELF", 700000)}); err == nil {
		t.Fatal("non-zero amount accepted on DISABLED_SELF")
	}
}

func TestFullScreeningGroupLeavesSeriousDiseaseTheRest(t *testing.T) {
	p := YA2025Draft()
	r, err := Evaluate(p, []Claim{
		claimFor(t, p, "a", "MEDICAL_SCREENING_EXAM", 100000),
		claimFor(t, p, "b", "MEDICAL_SCREENING_SELF_TEST_DEVICE", 5000),
		claimFor(t, p, "c", "MEDICAL_SERIOUS_DISEASE", 1000000),
	})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{100000, 0, 900000}, 1000000)
}

func TestSameExpenseCannotBeClaimedUnderParentsAndEquipment(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "PARENTS_MEDICAL", 200000)
	b := claimFor(t, p, "b", "DISABILITY_EQUIPMENT", 200000)
	b.ExpenseKey = a.ExpenseKey
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{200000, 0}, 200000)
	if !hasReason(r.Decisions[1], "DUPLICATE_EXPENSE") {
		t.Fatalf("%v", r.Decisions[1].Reasons)
	}
}

func TestMedicalFailsClosedOnUnverifiedFacts(t *testing.T) {
	cases := []struct {
		rule, fact string
		amt        int64
	}{
		{"DISABILITY_EQUIPMENT", "disability_certified_by_jkm", 100000},
		{"DISABLED_SELF", "disability_certified_by_jkm", 0},
		{"MEDICAL_SERIOUS_DISEASE", "qualifying_serious_disease", 100000}, // disease not named in PR §6.6.2
		{"PARENTS_CHECKUP", "qualifying_examination", 50000},              // laboratory test only, PR §6.7.4
		{"MEDICAL_SCREENING_EXAM", "qualifying_examination", 50000},
		{"MEDICAL_LEARNING_DISABILITY", "child_age_18_or_below", 100000},
		{"MEDICAL_FERTILITY", "qualifying_beneficiary", 100000}, // not self or spouse
	}
	p := YA2025Draft()
	for _, tc := range cases {
		t.Run(tc.rule+"/"+tc.fact, func(t *testing.T) {
			c := claimFor(t, p, "a", tc.rule, tc.amt)
			if _, ok := c.Facts[tc.fact]; !ok {
				t.Fatalf("%s does not require %s", tc.rule, tc.fact)
			}
			c.Facts[tc.fact] = false
			r, err := Evaluate(p, []Claim{c})
			if err != nil {
				t.Fatal(err)
			}
			d := r.Decisions[0]
			if d.Status != "REVIEW_REQUIRED" || d.ProvisionalAmountSen != 0 || !hasReason(d, "UNVERIFIED_FACT:"+tc.fact) {
				t.Fatalf("%+v", d)
			}
		})
	}
}
