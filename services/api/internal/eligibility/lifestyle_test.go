package eligibility

import (
	"fmt"
	"testing"
)

// Lifestyle (YA2025 item 9): RM2,500 shared cap across books, device, internet and course.
// These tests check draft arithmetic and fail-closed behaviour only. They do not
// prove that the draft rules match the statute.

const lifestyleCap = int64(250000)

var lifestyleRules = []string{"LIFESTYLE_BOOKS", "LIFESTYLE_DEVICE", "LIFESTYLE_INTERNET", "LIFESTYLE_COURSE"}

func ruleByID(t *testing.T, id string) Rule {
	t.Helper()
	for _, r := range Draft2025().Rules {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("rule %s missing from draft catalog", id)
	return Rule{}
}

func TestLifestyleSingleClaimCapBoundaries(t *testing.T) {
	cases := []struct {
		name         string
		amount, want int64
		limitReached bool
	}{
		{"one sen", 1, 1, false},
		{"cap minus one sen", lifestyleCap - 1, lifestyleCap - 1, false},
		{"exactly cap", lifestyleCap, lifestyleCap, false},
		{"cap plus one sen", lifestyleCap + 1, lifestyleCap, true},
		{"far above cap", lifestyleCap * 10, lifestyleCap, true},
	}
	for _, rule := range lifestyleRules {
		for _, tc := range cases {
			t.Run(rule+"/"+tc.name, func(t *testing.T) {
				r, err := Evaluate(Draft2025(), []Claim{claim("a", "e1", rule, tc.amount)})
				if err != nil {
					t.Fatal(err)
				}
				d := r.Decisions[0]
				if d.CandidateSen != tc.want || r.ProvisionalTotalSen != tc.want {
					t.Fatalf("candidate=%d total=%d want=%d", d.CandidateSen, r.ProvisionalTotalSen, tc.want)
				}
				has := false
				for _, reason := range d.Reasons {
					if reason == "LIMIT_REACHED" {
						has = true
					}
				}
				if has != tc.limitReached {
					t.Fatalf("LIMIT_REACHED=%v want %v reasons=%v", has, tc.limitReached, d.Reasons)
				}
			})
		}
	}
}

func TestLifestyleCapIsSharedAcrossAllFourCategories(t *testing.T) {
	claims := []Claim{
		claim("a", "e1", "LIFESTYLE_BOOKS", 60000),
		claim("b", "e2", "LIFESTYLE_INTERNET", 90000),
		claim("c", "e3", "LIFESTYLE_DEVICE", 80000),
		claim("d", "e4", "LIFESTYLE_COURSE", 70000),
	}
	r, err := Evaluate(Draft2025(), claims)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{60000, 90000, 80000, 20000} // allocated in claim ID order
	for i, d := range r.Decisions {
		if d.CandidateSen != want[i] {
			t.Fatalf("claim %s got %d want %d", d.ID, d.CandidateSen, want[i])
		}
	}
	if r.ProvisionalTotalSen != lifestyleCap {
		t.Fatalf("total=%d want %d", r.ProvisionalTotalSen, lifestyleCap)
	}
}

func TestLifestyleAllocationIsOrderIndependent(t *testing.T) {
	base := []Claim{
		claim("a", "e1", "LIFESTYLE_BOOKS", 100000),
		claim("b", "e2", "LIFESTYLE_DEVICE", 100000),
		claim("c", "e3", "LIFESTYLE_INTERNET", 100000),
	}
	perms := [][]int{{0, 1, 2}, {2, 1, 0}, {1, 2, 0}, {1, 0, 2}}
	var first map[string]int64
	for _, p := range perms {
		in := []Claim{base[p[0]], base[p[1]], base[p[2]]}
		r, err := Evaluate(Draft2025(), in)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]int64{}
		for _, d := range r.Decisions {
			got[d.ID] = d.CandidateSen
		}
		if first == nil {
			first = got
			if got["a"] != 100000 || got["b"] != 100000 || got["c"] != 50000 {
				t.Fatalf("unexpected allocation %v", got)
			}
			continue
		}
		for k, v := range first {
			if got[k] != v {
				t.Fatalf("order changed allocation: %v vs %v", got, first)
			}
		}
	}
}

func TestLifestyleEachRequiredFactBlocksClaim(t *testing.T) {
	for _, rule := range lifestyleRules {
		for _, fact := range ruleByID(t, rule).RequiredFacts {
			t.Run(rule+"/missing "+fact, func(t *testing.T) {
				bad := claim("a", "e1", rule, 100000)
				delete(bad.Facts, fact)
				good := claim("b", "e2", "LIFESTYLE_BOOKS", lifestyleCap)
				r, err := Evaluate(Draft2025(), []Claim{bad, good})
				if err != nil {
					t.Fatal(err)
				}
				if r.Decisions[0].CandidateSen != 0 || r.Decisions[0].State != "REVIEW_REQUIRED" {
					t.Fatalf("unverified fact must block: %+v", r.Decisions[0])
				}
				found := false
				for _, reason := range r.Decisions[0].Reasons {
					if reason == "FACT_NOT_VERIFIED:"+fact {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing reason for %s: %v", fact, r.Decisions[0].Reasons)
				}
				// A blocked claim must not consume the shared cap.
				if r.Decisions[1].CandidateSen != lifestyleCap {
					t.Fatalf("blocked claim consumed cap, next got %d", r.Decisions[1].CandidateSen)
				}
			})
		}
	}
}

func TestLifestyleFalseFactBlocksLikeMissing(t *testing.T) {
	c := claim("a", "e1", "LIFESTYLE_DEVICE", 100000)
	c.Facts["not_business_use"] = false
	r, err := Evaluate(Draft2025(), []Claim{c})
	if err != nil || r.ProvisionalTotalSen != 0 {
		t.Fatalf("business-use device accepted: %+v %v", r, err)
	}
}

func TestLifestyleEvidenceAndExpenseIdentifierRequired(t *testing.T) {
	for _, rule := range lifestyleRules {
		c := claim("a", "e1", rule, 100000)
		c.EvidenceReference = ""
		r, _ := Evaluate(Draft2025(), []Claim{c})
		if r.ProvisionalTotalSen != 0 || r.Decisions[0].Reasons[0] != "EVIDENCE_MISSING" {
			t.Fatalf("%s: missing evidence must block: %+v", rule, r.Decisions[0])
		}
		c = claim("a", "", rule, 100000)
		r, _ = Evaluate(Draft2025(), []Claim{c})
		if r.ProvisionalTotalSen != 0 || r.Decisions[0].Reasons[0] != "EXPENSE_IDENTIFIER_MISSING" {
			t.Fatalf("%s: missing expense id must block: %+v", rule, r.Decisions[0])
		}
	}
}

func TestLifestyleSameExpenseCannotBeClaimedTwice(t *testing.T) {
	// Same expense in two lifestyle categories.
	r, err := Evaluate(Draft2025(), []Claim{
		claim("a", "inv-1", "LIFESTYLE_DEVICE", 100000),
		claim("b", "inv-1", "LIFESTYLE_COURSE", 100000),
	})
	if err != nil || r.ProvisionalTotalSen != 100000 || r.Decisions[1].CandidateSen != 0 {
		t.Fatalf("double count within lifestyle: %+v %v", r, err)
	}
	// Same course fee claimed as lifestyle and as education up-skilling.
	r, err = Evaluate(Draft2025(), []Claim{
		claim("a", "course-9", "EDUCATION_UPSKILL", 100000),
		claim("b", "course-9", "LIFESTYLE_COURSE", 100000),
	})
	if err != nil || r.ProvisionalTotalSen != 100000 {
		t.Fatalf("course fee counted in lifestyle and education: %+v %v", r, err)
	}
	dup := false
	for _, reason := range r.Decisions[1].Reasons {
		if reason == "DUPLICATE_EXPENSE" {
			dup = true
		}
	}
	if !dup {
		t.Fatalf("expected DUPLICATE_EXPENSE, got %v", r.Decisions[1].Reasons)
	}
}

func TestLifestylePoolIsSeparateFromMedicalAndEducation(t *testing.T) {
	r, err := Evaluate(Draft2025(), []Claim{
		claim("a", "e1", "LIFESTYLE_BOOKS", lifestyleCap),
		claim("b", "e2", "MEDICAL_VACCINATION", 100000),
		claim("c", "e3", "EDUCATION_UPSKILL", 200000),
	})
	if err != nil || r.ProvisionalTotalSen != lifestyleCap+100000+200000 {
		t.Fatalf("pools leaked into each other: %+v %v", r, err)
	}
}

func TestLifestyleNeverApprovedOrFilingReady(t *testing.T) {
	for _, rule := range lifestyleRules {
		r, err := Evaluate(Draft2025(), []Claim{claim("a", "e1", rule, 100000)})
		if err != nil {
			t.Fatal(err)
		}
		d := r.Decisions[0]
		draft := false
		for _, reason := range d.Reasons {
			if reason == "DRAFT_LEGAL_RULE_NOT_APPROVED" {
				draft = true
			}
		}
		if r.ApprovedTotalSen != 0 || r.FilingReady || d.State != "REVIEW_REQUIRED" || !draft {
			t.Fatalf("%s leaked approval: %+v", rule, r)
		}
	}
}

func TestLifestyleRejectsBadInputAndUnsupportedYear(t *testing.T) {
	for _, amt := range []int64{0, -1} {
		if _, err := Evaluate(Draft2025(), []Claim{claim("a", "e1", "LIFESTYLE_BOOKS", amt)}); err == nil {
			t.Fatalf("amount %d accepted", amt)
		}
	}
	if _, err := Evaluate(Draft2025(), []Claim{claim("a", "e1", "LIFESTYLE_BOOKS", 1), claim("a", "e2", "LIFESTYLE_BOOKS", 1)}); err == nil {
		t.Fatal("duplicate claim ID accepted")
	}
	for _, year := range []int{2023, 2024, 2026} {
		cat := Draft2025()
		cat.Year = year
		if _, err := Evaluate(cat, []Claim{claim("a", "e1", "LIFESTYLE_BOOKS", 1)}); err == nil {
			t.Fatalf("year %d accepted", year)
		}
	}
	r, _ := Evaluate(Draft2025(), []Claim{{ID: "a", ExpenseKey: "e1", RuleID: "LIFESTYLE_UNKNOWN", AmountSen: 1, EvidenceReference: "x"}})
	if r.ProvisionalTotalSen != 0 || r.Decisions[0].Reasons[0] != "UNKNOWN_RULE" {
		t.Fatalf("unknown rule must fail closed: %+v", r.Decisions[0])
	}
}

func TestLifestyleManySmallClaimsStopExactlyAtCap(t *testing.T) {
	var claims []Claim
	for i := 0; i < 26; i++ { // 26 x RM100 = RM2,600
		claims = append(claims, claim(fmt.Sprintf("c%02d", i), fmt.Sprintf("e%02d", i), "LIFESTYLE_BOOKS", 10000))
	}
	r, err := Evaluate(Draft2025(), claims)
	if err != nil || r.ProvisionalTotalSen != lifestyleCap {
		t.Fatalf("total=%d err=%v", r.ProvisionalTotalSen, err)
	}
	if r.Decisions[24].CandidateSen != 10000 || r.Decisions[25].CandidateSen != 0 {
		t.Fatalf("cap edge wrong: %d %d", r.Decisions[24].CandidateSen, r.Decisions[25].CandidateSen)
	}
}
