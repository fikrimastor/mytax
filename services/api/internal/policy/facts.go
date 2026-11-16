package policy

import "sort"

// CanonicalFacts is the one list of fact names a rule may require.
// Conventions: snake_case; qualifying_<thing> describes the item or service;
// <thing>_in_own_name describes ownership; resident, within_year, no_other_claim
// and qualifying_beneficiary are shared across rules. Keep the list sorted.
var CanonicalFacts = []string{
	"allied_health_practitioner_registered",
	"child_age_18_or_below",
	"citizen",
	"co_claimant_none",
	"disability_certified_by_jkm",
	"first_home",
	"individual_married",
	"internet_in_own_name",
	"medical_practitioner_certified",
	"no_income_from_property",
	"no_other_claim",
	"not_business_use",
	"not_course_under_46_1_f_iii",
	"occupied_as_residence",
	"own_education",
	"paid_by_claimant",
	"parent_resident_in_malaysia",
	"price_above_500k_upto_750k",
	"price_up_to_500k",
	"qualifying_beneficiary",
	"qualifying_carer",
	"qualifying_course",
	"qualifying_dental",
	"qualifying_detection_test",
	"qualifying_device",
	"qualifying_examination",
	"qualifying_expense",
	"qualifying_fertility_treatment",
	"qualifying_interest",
	"qualifying_learning_disability",
	"qualifying_medical_treatment",
	"qualifying_mental_health_provider",
	"qualifying_reading_material",
	"qualifying_self_test_device",
	"qualifying_serious_disease",
	"qualifying_spa_date",
	"qualifying_supporting_equipment",
	"qualifying_vaccine",
	"resident",
	"treatment_in_malaysia",
	"upskilling_course",
	"within_three_years_of_spa",
	"within_year",
}

// IsCanonicalFact reports whether name is in CanonicalFacts.
func IsCanonicalFact(name string) bool {
	i := sort.SearchStrings(CanonicalFacts, name)
	return i < len(CanonicalFacts) && CanonicalFacts[i] == name
}
