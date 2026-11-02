package policy

// Act53URL is the HASiL-hosted PDF of the Income Tax Act 1967 supplied by the repository
// owner. The text used is the AGC "updated text as at 21 May 2024", NOT an authentic text.
// See docs/sources/act53 for the provenance, hashes and excerpt.
const Act53URL = "https://www.hasil.gov.my/wp-content/uploads/20240521-akta-cukai-pendapatan-1967-akta-53.pdf"

// lifestyleSources cites Act 53 s.46(1)(p) paragraph para and the user-supplied LHDN table.
// Both Act texts are AGC updated texts, not authentic texts; (p) is identical in the two.
func lifestyleSources(para string) []Source {
	return []Source{
		{URL: Act53URL, Section: "Income Tax Act 1967 (Act 53) updated texts as at 21 May 2024 and 1 January 2026, s.46(1)(p)" + para + " and s.46(3); not an authentic text; paragraph (p) reads the same in both", RetrievedOn: "2026-11-02"},
		{URL: "https://www.hasil.gov.my/individu/pelepasan-cukai/", Section: "Tahun Taksiran 2025, item 9 Gaya hidup (user-supplied snapshot)", RetrievedOn: "2026-11-02"},
	}
}

// firstHomeSources cites Act 53 s.46B as replaced by Finance Act 2024 s.10, in force
// 1 January 2025. Both Act texts are AGC updated texts, not authentic texts.
func firstHomeSources() []Source {
	return []Source{
		{URL: Act53URL, Section: "Income Tax Act 1967 (Act 53) updated texts as at 21 May 2024 and 1 January 2026, s.46B as replaced by Finance Act 2024 s.10 (in force 1 January 2025); not an authentic text", RetrievedOn: "2026-11-02"},
		{URL: "https://www.hasil.gov.my/individu/pelepasan-cukai/", Section: "Tahun Taksiran 2025, first residential property interest relief (user-supplied snapshot)", RetrievedOn: "2026-11-02"},
	}
}

// YA2025Draft is a deliberately INACTIVE subset of the official relief schedule.
// Group allocations represent accounting limits only; they do not prove entitlement.
// No production tax computation may depend on this draft catalog.
func YA2025Draft() Package {
	src := []Source{{URL: "https://www.hasil.gov.my/individu/pelepasan-cukai/", Section: "Tahun Taksiran 2025", RetrievedOn: "2026-10-31"}}
	return Package{AssessmentYear: 2025, Version: "ya2025-draft-001", Status: Draft,
		Groups: []Group{{ID: "LIFESTYLE", LimitSen: 250000}, {ID: "MEDICAL_SELF_FAMILY", LimitSen: 1000000}, {ID: "EDUCATION_SELF", LimitSen: 700000}},
		Rules: []Rule{
			{ID: "LIFESTYLE_INTERNET", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "internet_in_own_name", "qualifying_expense", "no_other_claim"}, Sources: lifestyleSources("(iv)")},
			{ID: "LIFESTYLE_BOOKS", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "qualifying_reading_material", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(i)")},
			{ID: "LIFESTYLE_DEVICE", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "qualifying_device", "not_business_use", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(ii)")},
			// s.46(1)(p)(v): a course for upskilling or self-enhancement OTHER THAN a course falling under s.46(1)(f)(iii).
			{ID: "LIFESTYLE_COURSE", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "upskilling_course", "not_course_under_46_1_f_iii", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(v)")},
			{ID: "MEDICAL_VACCINATION", Group: "MEDICAL_SELF_FAMILY", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_vaccine", "qualifying_beneficiary", "no_other_claim"}, Sources: src},
			{ID: "EDUCATION_SELF_SKILLS", Group: "EDUCATION_SELF", LimitSen: 200000, RequiredFacts: []string{"resident", "within_year", "qualifying_course", "own_education", "no_other_claim"}, Sources: src},
			{ID: "EDUCATION_SELF_OTHER", Group: "EDUCATION_SELF", LimitSen: 700000, RequiredFacts: []string{"resident", "within_year", "qualifying_course", "own_education", "no_other_claim"}, Sources: src},
			// s.46B first-home interest. Fact names mirror the eligibility engine's
			// FIRST_HOME_* rules; internal/eligibility/catalog_parity_test.go enforces parity.
			{ID: "FIRST_HOME_LOW", LimitSen: 700000, RequiredFacts: []string{"resident", "citizen", "first_home", "occupied_as_residence", "no_income_from_property", "qualifying_spa_date", "price_up_to_500k", "qualifying_interest"}, Sources: firstHomeSources()},
			{ID: "FIRST_HOME_MID", LimitSen: 500000, RequiredFacts: []string{"resident", "citizen", "first_home", "occupied_as_residence", "no_income_from_property", "qualifying_spa_date", "price_above_500k_upto_750k", "qualifying_interest"}, Sources: firstHomeSources()},
		}}
}
