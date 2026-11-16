package policy

import "fmt"

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

// PR72025URL is the official IRBM URL of LHDN Public Ruling No. 7/2025 (7th edition, 5 December 2025).
// It returned HTTP 500 on 2026-11-07 and 2026-11-16; the text used is the MAICSA mirror recorded,
// with its SHA-256, in docs/sources/pr7-2025. A ruling is an administrative interpretation (s.138A), not law.
const PR72025URL = "https://www.hasil.gov.my/wp-content/uploads/pr-7-2025.pdf"

// medicalSources cites the Act 53 provision for YA 2025 and the PR 7/2025 paragraph that interprets it.
func medicalSources(act, ruling string) []Source {
	return []Source{
		{URL: Act53URL, Section: "Income Tax Act 1967 (Act 53) updated text as at 21 May 2024, " + act + "; not an authentic text", RetrievedOn: "2026-11-07"},
		{URL: PR72025URL, Section: "LHDN Public Ruling No. 7/2025 (7th edition, 5 December 2025) " + ruling + "; read from the MAICSA mirror (docs/sources/pr7-2025) because the IRBM URL returned HTTP 500", RetrievedOn: "2026-11-16"},
	}
}

// YA2025Draft is a deliberately INACTIVE subset of the official relief schedule.
// Group allocations represent accounting limits only; they do not prove entitlement.
// No production tax computation may depend on this draft catalog.
func YA2025Draft() Package {
	src := []Source{{URL: "https://www.hasil.gov.my/individu/pelepasan-cukai/", Section: "Tahun Taksiran 2025", RetrievedOn: "2026-10-31"}}
	return Package{AssessmentYear: 2025, Version: "ya2025-draft-001", Status: Draft,
		Groups: []Group{
			{ID: "LIFESTYLE", LimitSen: 250000},
			{ID: "MEDICAL_SELF_FAMILY", LimitSen: 1000000},
			// s.46(1)(h): RM1,000 within the RM10,000 of s.46(1)(g).
			{ID: "MEDICAL_SCREENING", Parent: "MEDICAL_SELF_FAMILY", LimitSen: 100000},
			{ID: "EDUCATION_SELF", LimitSen: 700000},
			// s.46(1)(c): one RM8,000 cap for all parents and grandparents together (interpretation; PR 7/2025 Example 3).
			{ID: "PARENTS", LimitSen: 800000},
			{ID: "DISABILITY_EQUIPMENT", LimitSen: 600000},
		},
		Rules: []Rule{
			{ID: "LIFESTYLE_INTERNET", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "internet_in_own_name", "qualifying_expense", "no_other_claim"}, Sources: lifestyleSources("(iv)")},
			{ID: "LIFESTYLE_BOOKS", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "qualifying_reading_material", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(i)")},
			{ID: "LIFESTYLE_DEVICE", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "qualifying_device", "not_business_use", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(ii)")},
			// s.46(1)(p)(v): a course for upskilling or self-enhancement OTHER THAN a course falling under s.46(1)(f)(iii).
			{ID: "LIFESTYLE_COURSE", Group: "LIFESTYLE", LimitSen: 250000, RequiredFacts: []string{"resident", "within_year", "upskilling_course", "not_course_under_46_1_f_iii", "qualifying_beneficiary", "no_other_claim"}, Sources: lifestyleSources("(v)")},
			{ID: "MEDICAL_VACCINATION", Group: "MEDICAL_SELF_FAMILY", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_vaccine", "qualifying_beneficiary", "no_other_claim"}, Sources: src},
			{ID: "MEDICAL_DENTAL", Group: "MEDICAL_SELF_FAMILY", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_dental", "qualifying_beneficiary", "no_other_claim"}, Sources: src},
			// s.46(1)(c) parents and grandparents. No apportionment between siblings: each claimant claims what they paid.
			{ID: "PARENTS_MEDICAL", Group: "PARENTS", LimitSen: 800000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_medical_treatment", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c)(i) as amended by Finance Act 2024 s.9(a); RM8,000 is read as one cap for all parents and grandparents (interpretation)", "paragraphs 6.2.1-6.2.5, Example 3 and Appendix")},
			{ID: "PARENTS_CARER", Group: "PARENTS", LimitSen: 800000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_carer", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c)(ii) as amended by Finance Act 2024 s.9(a)", "paragraphs 6.2.1-6.2.4")},
			{ID: "PARENTS_CHECKUP", Group: "PARENTS", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_examination", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c) proviso (e) as amended by Finance Act 2024 s.9(a): complete medical examination including any vaccination, up to RM1,000", "paragraph 6.2.3(c) and Example 3")},
			// s.46(1)(d). One expense cannot also be claimed under (c): the shared expense_key blocks it (DUPLICATE_EXPENSE).
			{ID: "DISABILITY_EQUIPMENT", Group: "DISABILITY_EQUIPMENT", LimitSen: 600000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "disability_certified_by_jkm", "qualifying_supporting_equipment", "no_other_claim"},
				Sources: medicalSources("s.46(1)(d)", "paragraphs 3.6, 6.3.1-6.3.2 and Example 4")},
			// s.46(1)(e): a fixed RM7,000 for a disabled individual; claim with amount_sen 0.
			{ID: "DISABLED_SELF", FixedAmount: true, LimitSen: 700000, RequiredFacts: []string{"resident", "within_year", "disability_certified_by_jkm"},
				Sources: medicalSources("s.46(1)(e) as amended by Finance Act 2024 s.9(b) (RM6,000 to RM7,000 from YA 2025)", "paragraphs 3.6 and 6.4")},
			{ID: "MEDICAL_SERIOUS_DISEASE", Group: "MEDICAL_SELF_FAMILY", LimitSen: 1000000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "qualifying_serious_disease", "no_other_claim"},
				Sources: medicalSources("s.46(1)(g)(i); serious disease read as the diseases named in PR 7/2025 paragraph 6.6.2 (interpretation)", "paragraphs 6.6.1-6.6.2 and Examples 8, 11, 13, 16, 17, 19")},
			{ID: "MEDICAL_FERTILITY", Group: "MEDICAL_SELF_FAMILY", LimitSen: 1000000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "qualifying_fertility_treatment", "individual_married", "no_other_claim"},
				Sources: medicalSources("s.46(1)(g)(ii)", "paragraph 6.6.3 and Examples 9, 10, 11")},
			{ID: "MEDICAL_SCREENING_EXAM", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_examination", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(i) as amended by Finance Act 2024 s.9(c)", "paragraphs 6.7.1-6.7.4 and Example 14")},
			{ID: "MEDICAL_SCREENING_DETECTION_TEST", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_detection_test", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(i) as amended by Finance Act 2024 s.9(c)", "paragraph 6.7.6 and Examples 16, 17")},
			{ID: "MEDICAL_SCREENING_SELF_TEST_DEVICE", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_self_test_device", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(ii) as amended by Finance Act 2024 s.9(c): registered under the Medical Device Act 2012 (Act 737), not for business use", "paragraph 6.7.7 and Example 17")},
			{ID: "MEDICAL_SCREENING_MENTAL_HEALTH", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_mental_health_provider", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(iii)", "paragraph 6.7.5 and Example 14")},
			{ID: "MEDICAL_LEARNING_DISABILITY", Group: "MEDICAL_SELF_FAMILY", LimitSen: 600000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "treatment_in_malaysia", "child_age_18_or_below", "qualifying_learning_disability", "allied_health_practitioner_registered", "no_other_claim"},
				Sources: medicalSources("s.46(1)(ha) as amended by Finance Act 2024 s.9(d) (RM6,000 for YA 2025; RM10,000 starts YA 2026)", "paragraphs 6.8.1-6.8.3 and Examples 18, 19, 20")},
			{ID: "EDUCATION_SELF_SKILLS", Group: "EDUCATION_SELF", LimitSen: 200000, RequiredFacts: []string{"resident", "within_year", "qualifying_course", "own_education", "no_other_claim"}, Sources: src},
			{ID: "EDUCATION_SELF_OTHER", Group: "EDUCATION_SELF", LimitSen: 700000, RequiredFacts: []string{"resident", "within_year", "qualifying_course", "own_education", "no_other_claim"}, Sources: src},
			// s.46B first-home interest. The two price tiers are mutually exclusive (ExclusiveFamily).
			{ID: "FIRST_HOME_LOW", ExclusiveFamily: "FIRST_HOME", LimitSen: 700000, RequiredFacts: []string{"resident", "citizen", "first_home", "occupied_as_residence", "no_income_from_property", "qualifying_spa_date", "within_three_years_of_spa", "co_claimant_none", "price_up_to_500k", "qualifying_interest"}, Sources: firstHomeSources()},
			{ID: "FIRST_HOME_MID", ExclusiveFamily: "FIRST_HOME", LimitSen: 500000, RequiredFacts: []string{"resident", "citizen", "first_home", "occupied_as_residence", "no_income_from_property", "qualifying_spa_date", "within_three_years_of_spa", "co_claimant_none", "price_above_500k_upto_750k", "qualifying_interest"}, Sources: firstHomeSources()},
		}}
}

// DraftPackage returns the draft (unapproved) rule package for an assessment year.
// Only YA 2025 has one; every other year fails closed.
func DraftPackage(year int) (Package, error) {
	if year != supportedYear {
		return Package{}, fmt.Errorf("no draft rule package for assessment year %d", year)
	}
	return YA2025Draft(), nil
}
