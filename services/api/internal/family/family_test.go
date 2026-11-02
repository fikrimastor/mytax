package family

import "testing"

func okChild(id, cat string) Child {
	c := Child{ID: id, Category: cat, Unmarried: true, TotalIncomeVerified: true}
	switch cat {
	case CatUnder18:
		c.UnderEighteenAtAnyTime = true
	case CatInstruction:
		c.FullTimeInstruction = true
	case CatArticles:
		c.ServingArticlesOrIndentures = true
	case CatDisabled:
		c.DisabilityProvenToDG = true
	}
	return c
}
func claim(id, who, child string) ChildClaim {
	return ChildClaim{ID: id, ClaimantID: who, ChildID: child, ClaimantResident: true, PaidMaintenance: true}
}
func byID(r Result, id string) Decision {
	for _, d := range r.Decisions {
		if d.ID == id {
			return d
		}
	}
	return Decision{ID: "missing"}
}
func has(d Decision, reason string) bool {
	for _, x := range d.Reasons {
		if x == reason {
			return true
		}
	}
	return false
}
func eval(t *testing.T, kids []Child, claims []ChildClaim) Result {
	t.Helper()
	r, err := ChildDeductions(kids, claims)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestChildSingleClaimantAmounts(t *testing.T) {
	cases := []struct {
		name string
		kid  func() Child
		want int64
	}{
		{"under 18", func() Child { return okChild("k", CatUnder18) }, 200000},
		{"over 18 at school", func() Child { c := okChild("k", CatInstruction); c.Over18 = true; return c }, 200000},
		{"higher education approved", func() Child {
			c := okChild("k", CatInstruction)
			c.Over18, c.HigherEducationOrArticles, c.ApprovedByAuthority = true, true, true
			return c
		}, 800000},
		{"articles over 18 approved", func() Child {
			c := okChild("k", CatArticles)
			c.Over18, c.HigherEducationOrArticles, c.ApprovedByAuthority = true, true, true
			return c
		}, 800000},
		{"abroad with degree", func() Child {
			c := okChild("k", CatInstruction)
			c.Over18, c.HigherEducationOrArticles, c.ApprovedByAuthority, c.OutsideMalaysia, c.DegreeOrEquivalent = true, true, true, true, true
			return c
		}, 800000},
		{"disabled", func() Child { return okChild("k", CatDisabled) }, 800000},
		{"disabled in higher education", func() Child {
			c := okChild("k", CatDisabled)
			c.Over18, c.HigherEducationOrArticles, c.ApprovedByAuthority = true, true, true
			return c
		}, 1600000},
	}
	for _, tc := range cases {
		r := eval(t, []Child{tc.kid()}, []ChildClaim{claim("c1", "p1", "k")})
		d := byID(r, "c1")
		if d.CandidateSen != tc.want || r.CandidateTotalSen != tc.want {
			t.Fatalf("%s: got %d want %d reasons=%v", tc.name, d.CandidateSen, tc.want, d.Reasons)
		}
		if r.ApprovedTotalSen != 0 || r.FilingReady || d.State != "REVIEW_REQUIRED" || !has(d, "DRAFT_LEGAL_RULE_NOT_APPROVED") {
			t.Fatalf("%s: leaked approval: %+v", tc.name, r)
		}
	}
}

func TestChildUpliftConditionsFailClosed(t *testing.T) {
	mk := func(mod func(*Child)) Child {
		c := okChild("k", CatInstruction)
		c.Over18, c.HigherEducationOrArticles, c.ApprovedByAuthority = true, true, true
		mod(&c)
		return c
	}
	for name, kid := range map[string]Child{
		"not approved":       mk(func(c *Child) { c.ApprovedByAuthority = false }),
		"abroad no degree":   mk(func(c *Child) { c.OutsideMalaysia = true }),
		"married":            mk(func(c *Child) { c.Unmarried = false }),
		"instruction absent": mk(func(c *Child) { c.FullTimeInstruction = false }),
	} {
		r := eval(t, []Child{kid}, []ChildClaim{claim("c1", "p1", "k")})
		if r.CandidateTotalSen != 0 || byID(r, "c1").State != "REVIEW_REQUIRED" {
			t.Fatalf("%s must block: %+v", name, r)
		}
	}
	d := okChild("k", CatDisabled)
	d.DisabilityProvenToDG = false
	if r := eval(t, []Child{d}, []ChildClaim{claim("c1", "p1", "k")}); r.CandidateTotalSen != 0 {
		t.Fatalf("unproven disability must block: %+v", r)
	}
	if r := eval(t, []Child{{ID: "k", Category: "OTHER", Unmarried: true, TotalIncomeVerified: true}}, []ChildClaim{claim("c1", "p1", "k")}); r.CandidateTotalSen != 0 || !has(byID(r, "c1"), "UNKNOWN_CHILD_CATEGORY") {
		t.Fatalf("unknown category must block: %+v", r)
	}
}

func TestTwoClaimantsGetFiftyPercentEach(t *testing.T) {
	r := eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{claim("b", "p2", "k"), claim("a", "p1", "k")})
	if byID(r, "a").CandidateSen != 100000 || byID(r, "b").CandidateSen != 100000 || r.CandidateTotalSen != 200000 {
		t.Fatalf("%+v", r)
	}
	r = eval(t, []Child{okChild("k", CatDisabled)}, []ChildClaim{claim("a", "p1", "k"), claim("b", "p2", "k")})
	if byID(r, "a").CandidateSen != 400000 {
		t.Fatalf("%+v", r)
	}
}

func TestMoreThanTwoClaimantsFailClosed(t *testing.T) {
	r := eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{claim("a", "p1", "k"), claim("b", "p2", "k"), claim("c", "p3", "k")})
	if r.CandidateTotalSen != 0 || !has(byID(r, "a"), "MORE_THAN_TWO_CLAIMANTS_UNRESOLVED") {
		t.Fatalf("%+v", r)
	}
}

func TestSameClaimantTwiceForSameChildCountsOnce(t *testing.T) {
	r := eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{claim("a", "p1", "k"), claim("b", "p1", "k")})
	if r.CandidateTotalSen != 200000 || !has(byID(r, "b"), "DUPLICATE_CLAIM_FOR_SAME_CHILD_BY_SAME_CLAIMANT") {
		t.Fatalf("%+v", r)
	}
}

func TestWifeElectionWholeDeduction(t *testing.T) {
	wife := claim("w", "wife", "k")
	wife.WifeElectsWhole, wife.WifeLivingWithHusbandAssessedSeparately = true, true
	r := eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{claim("h", "husband", "k"), wife})
	if byID(r, "w").CandidateSen != 200000 || byID(r, "h").CandidateSen != 0 || !has(byID(r, "h"), "WIFE_ELECTION_GIVES_WHOLE_DEDUCTION_TO_OTHER_CLAIMANT") || r.CandidateTotalSen != 200000 {
		t.Fatalf("%+v", r)
	}
	wife.WifeLivingWithHusbandAssessedSeparately = false
	r = eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{claim("h", "husband", "k"), wife})
	if r.CandidateTotalSen != 0 || !has(byID(r, "w"), "ELECTION_CONDITION_NOT_VERIFIED") {
		t.Fatalf("invalid election must block both: %+v", r)
	}
	solo := claim("w", "wife", "k")
	solo.WifeElectsWhole, solo.WifeLivingWithHusbandAssessedSeparately = true, true
	if r = eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{solo}); r.CandidateTotalSen != 0 {
		t.Fatalf("election without a second claimant must block: %+v", r)
	}
	h := claim("h", "husband", "k")
	h.WifeElectsWhole, h.WifeLivingWithHusbandAssessedSeparately = true, true
	w2 := solo
	if r = eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{h, w2}); r.CandidateTotalSen != 0 || !has(byID(r, "h"), "ELECTION_AMBIGUOUS") {
		t.Fatalf("two elections must block: %+v", r)
	}
}

func TestChildIncomeTestBoundaries(t *testing.T) {
	mk := func(income int64, verified bool) Child {
		c := okChild("k", CatUnder18)
		c.TotalIncomeSen, c.TotalIncomeVerified = income, verified
		return c
	}
	one := []ChildClaim{claim("a", "p1", "k")}
	two := []ChildClaim{claim("a", "p1", "k"), claim("b", "p2", "k")}
	cases := []struct {
		name   string
		kid    Child
		claims []ChildClaim
		want   int64
		reason string
	}{
		{"income equals deduction", mk(200000, true), one, 200000, ""},
		{"income one sen above", mk(200001, true), one, 0, "CHILD_INCOME_EXCEEDS_DEDUCTION"},
		{"zero income", mk(0, true), one, 200000, ""},
		{"unverified income", mk(0, false), one, 0, "CHILD_TOTAL_INCOME_NOT_VERIFIED"},
		{"two claimants income equals share", mk(100000, true), two, 100000, ""},
		{"two claimants income between share and whole", mk(100001, true), two, 0, "CHILD_INCOME_TEST_BASIS_UNRESOLVED"},
		{"two claimants income equals whole", mk(200000, true), two, 0, "CHILD_INCOME_TEST_BASIS_UNRESOLVED"},
		{"two claimants income above whole", mk(200001, true), two, 0, "CHILD_INCOME_EXCEEDS_DEDUCTION"},
	}
	for _, tc := range cases {
		r := eval(t, []Child{tc.kid}, tc.claims)
		d := byID(r, "a")
		if d.CandidateSen != tc.want || (tc.reason != "" && !has(d, tc.reason)) {
			t.Fatalf("%s: got %d reasons=%v", tc.name, d.CandidateSen, d.Reasons)
		}
	}
}

func TestChildClaimantFactsAndUnknownChild(t *testing.T) {
	bad := claim("a", "p1", "k")
	bad.ClaimantResident = false
	good := claim("b", "p2", "k")
	r := eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{bad, good})
	// The non-resident claimant is excluded, so the other claimant is the only one and gets the whole.
	if byID(r, "a").CandidateSen != 0 || byID(r, "b").CandidateSen != 200000 {
		t.Fatalf("%+v", r)
	}
	nm := claim("a", "p1", "k")
	nm.PaidMaintenance = false
	if r = eval(t, []Child{okChild("k", CatUnder18)}, []ChildClaim{nm}); r.CandidateTotalSen != 0 {
		t.Fatalf("%+v", r)
	}
	if r = eval(t, nil, []ChildClaim{claim("a", "p1", "ghost")}); !has(byID(r, "a"), "UNKNOWN_CHILD") {
		t.Fatalf("%+v", r)
	}
}

func TestChildOrderIndependenceAndRejectedInput(t *testing.T) {
	kids := []Child{okChild("k1", CatUnder18), okChild("k2", CatDisabled)}
	a := []ChildClaim{claim("1", "p1", "k1"), claim("2", "p1", "k2"), claim("3", "p2", "k2")}
	r1 := eval(t, kids, a)
	r2 := eval(t, []Child{kids[1], kids[0]}, []ChildClaim{a[2], a[0], a[1]})
	if r1.CandidateTotalSen != r2.CandidateTotalSen || r1.CandidateTotalSen != 200000+800000 {
		t.Fatalf("%d %d", r1.CandidateTotalSen, r2.CandidateTotalSen)
	}
	for _, id := range []string{"1", "2", "3"} {
		if byID(r1, id).CandidateSen != byID(r2, id).CandidateSen {
			t.Fatalf("order changed claim %s", id)
		}
	}
	if _, err := ChildDeductions(kids, []ChildClaim{claim("x", "p", "k1"), claim("x", "p", "k2")}); err == nil {
		t.Fatal("duplicate claim ID accepted")
	}
	if _, err := ChildDeductions([]Child{kids[0], kids[0]}, nil); err == nil {
		t.Fatal("duplicate child accepted")
	}
	neg := okChild("k", CatUnder18)
	neg.TotalIncomeSen = -1
	if _, err := ChildDeductions([]Child{neg}, nil); err == nil {
		t.Fatal("negative income accepted")
	}
	if _, err := ChildDeductions(kids, []ChildClaim{{ID: "", ClaimantID: "p", ChildID: "k1"}}); err == nil {
		t.Fatal("empty claim ID accepted")
	}
}

// ---------------------------------------------------------------- spouse

func husband() SpouseInput {
	return SpouseInput{ID: "s", Role: RoleHusband, ClaimantResident: true, LivingTogether: true, ForeignIncomeVerified: true}
}
func wifeClaimant() SpouseInput {
	return SpouseInput{ID: "s", Role: RoleWife, ClaimantResident: true, S45AConditionMet: true, OnlyOneWifeClaimingS45A: true, ForeignIncomeVerified: true}
}
func spouse(t *testing.T, in SpouseInput) Decision {
	t.Helper()
	r, err := SpouseDeduction(in)
	if err != nil || len(r.Decisions) != 1 || r.ApprovedTotalSen != 0 || r.FilingReady {
		t.Fatalf("%+v %v", r, err)
	}
	return r.Decisions[0]
}

func TestSpouseHusbandAmounts(t *testing.T) {
	if d := spouse(t, husband()); d.CandidateSen != 400000 || d.State != "REVIEW_REQUIRED" {
		t.Fatalf("%+v", d)
	}
	in := husband()
	in.SpouseDisabled, in.DisabilityProven = true, true
	if d := spouse(t, in); d.CandidateSen != 1000000 {
		t.Fatalf("disabled wife: %+v", d)
	}
	in.DisabilityProven = false
	if d := spouse(t, in); d.CandidateSen != 400000 || !has(d, "FACT_NOT_VERIFIED:spouse_disability_proven") {
		t.Fatalf("unproven disability must not add RM6,000: %+v", d)
	}
}

func TestSpouseAlimonyAndSharedFourThousandCap(t *testing.T) {
	in := husband()
	in.AlimonyPaidSen = 100000
	if d := spouse(t, in); d.CandidateSen != 400000 {
		t.Fatalf("wife deduction plus alimony must stay at RM4,000: %+v", d)
	}
	sep := husband()
	sep.LivingTogether = false
	for alimony, want := range map[int64]int64{0: 0, 300000: 300000, 400000: 400000, 400001: 400000, 900000: 400000} {
		sep.AlimonyPaidSen = alimony
		if d := spouse(t, sep); d.CandidateSen != want {
			t.Fatalf("alimony=%d got %d want %d", alimony, d.CandidateSen, want)
		}
	}
}

func TestSpouseAssessedSeparatelyGivesNoS47_1Deduction(t *testing.T) {
	in := husband()
	in.SpouseAssessedSeparately = true
	d := spouse(t, in)
	if d.CandidateSen != 0 || !has(d, "S47_4_WIFE_ASSESSED_SEPARATELY_NO_S47_1_DEDUCTION") {
		t.Fatalf("%+v", d)
	}
	in.SpouseDisabled, in.DisabilityProven = true, true
	if d = spouse(t, in); d.CandidateSen != 0 {
		t.Fatalf("further disabled amount must not survive s.47(4): %+v", d)
	}
	in.AlimonyPaidSen = 250000
	if d = spouse(t, in); d.CandidateSen != 250000 {
		t.Fatalf("alimony under s.47(2) is separate from s.47(4): %+v", d)
	}
}

func TestSpouseForeignIncomeBoundary(t *testing.T) {
	for income, want := range map[int64]int64{0: 400000, 400000: 400000, 400001: 0} {
		in := husband()
		in.SpouseForeignIncomeSen = income
		if d := spouse(t, in); d.CandidateSen != want {
			t.Fatalf("income=%d got %d want %d", income, d.CandidateSen, want)
		}
	}
	in := husband()
	in.ForeignIncomeVerified = false
	if d := spouse(t, in); d.CandidateSen != 0 || !has(d, "SPOUSE_FOREIGN_INCOME_NOT_VERIFIED") {
		t.Fatalf("%+v", d)
	}
	in = husband()
	in.SpouseDisabled, in.DisabilityProven, in.SpouseForeignIncomeSen = true, true, 99999999
	if d := spouse(t, in); d.CandidateSen != 1000000 {
		t.Fatalf("disabled spouse is exempt from the foreign income test: %+v", d)
	}
}

func TestSpouseWifeClaimsHusbandUnderS45A(t *testing.T) {
	if d := spouse(t, wifeClaimant()); d.CandidateSen != 400000 {
		t.Fatalf("%+v", d)
	}
	in := wifeClaimant()
	in.SpouseDisabled, in.DisabilityProven = true, true
	if d := spouse(t, in); d.CandidateSen != 1000000 {
		t.Fatalf("%+v", d)
	}
	for _, mod := range []func(*SpouseInput){
		func(i *SpouseInput) { i.S45AConditionMet = false },
		func(i *SpouseInput) { i.OnlyOneWifeClaimingS45A = false },
		func(i *SpouseInput) { i.SpouseForeignIncomeSen = 400001 },
	} {
		in = wifeClaimant()
		mod(&in)
		if d := spouse(t, in); d.CandidateSen != 0 {
			t.Fatalf("s.45A condition must block: %+v", d)
		}
	}
}

func TestSpouseNonResidentAndBadInput(t *testing.T) {
	in := husband()
	in.ClaimantResident = false
	if d := spouse(t, in); d.CandidateSen != 0 {
		t.Fatalf("%+v", d)
	}
	in = husband()
	in.LivingTogether = false
	if d := spouse(t, in); d.CandidateSen != 0 || !has(d, "FACT_NOT_VERIFIED:wife_living_together") {
		t.Fatalf("%+v", d)
	}
	for _, bad := range []SpouseInput{{ID: "s", Role: "OTHER"}, {ID: "", Role: RoleHusband}, {ID: "s", Role: RoleHusband, AlimonyPaidSen: -1}, {ID: "s", Role: RoleWife, SpouseForeignIncomeSen: -1}} {
		if _, err := SpouseDeduction(bad); err == nil {
			t.Fatalf("bad input accepted: %+v", bad)
		}
	}
}
