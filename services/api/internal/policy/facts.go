package policy

import "sort"

// CanonicalFacts is the one list of fact names a rule may require.
// Conventions: snake_case; qualifying_<thing> describes the item or service;
// <thing>_in_own_name describes ownership; resident, within_year, no_other_claim
// and qualifying_beneficiary are shared across rules. Keep the list sorted.
var CanonicalFacts = []string{
	"citizen",
	"co_claimant_none",
	"first_home",
	"internet_in_own_name",
	"no_income_from_property",
	"no_other_claim",
	"not_business_use",
	"not_course_under_46_1_f_iii",
	"occupied_as_residence",
	"own_education",
	"price_above_500k_upto_750k",
	"price_up_to_500k",
	"qualifying_beneficiary",
	"qualifying_course",
	"qualifying_dental",
	"qualifying_device",
	"qualifying_expense",
	"qualifying_interest",
	"qualifying_reading_material",
	"qualifying_spa_date",
	"qualifying_vaccine",
	"resident",
	"upskilling_course",
	"within_three_years_of_spa",
	"within_year",
}

// IsCanonicalFact reports whether name is in CanonicalFacts.
func IsCanonicalFact(name string) bool {
	i := sort.SearchStrings(CanonicalFacts, name)
	return i < len(CanonicalFacts) && CanonicalFacts[i] == name
}
