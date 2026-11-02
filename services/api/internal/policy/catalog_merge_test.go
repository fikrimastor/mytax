package policy

import (
	"reflect"
	"sort"
	"testing"
)

func TestFirstHomeTiersShareAnExclusiveFamily(t *testing.T) {
	p := YA2025Draft()
	for _, id := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		if got := findRule(t, p, id).ExclusiveFamily; got != "FIRST_HOME" {
			t.Fatalf("%s exclusive family = %q, want FIRST_HOME", id, got)
		}
	}
}

func TestDentalRuleIsInTheMedicalGroup(t *testing.T) {
	r := findRule(t, YA2025Draft(), "MEDICAL_DENTAL")
	if r.Group != "MEDICAL_SELF_FAMILY" || r.LimitSen != 100000 {
		t.Fatalf("dental group=%s limit=%d", r.Group, r.LimitSen)
	}
	need := map[string]bool{}
	for _, f := range r.RequiredFacts {
		need[f] = true
	}
	for _, f := range []string{"resident", "within_year", "qualifying_dental", "qualifying_beneficiary", "no_other_claim"} {
		if !need[f] {
			t.Fatalf("dental rule lacks fact %s: %v", f, r.RequiredFacts)
		}
	}
	if len(r.Sources) == 0 {
		t.Fatal("dental rule has no source")
	}
}

// When the eligibility engine was merged into policy, the policy fact lists
// won (they are stricter) for these rules. Clients that sent the old fact
// names (incurred_in_ya, recognized_course, qualifying_education,
// qualifying_person) now get UNVERIFIED_FACT reasons. equivalence_test.go
// builds facts from the NEW catalog and therefore cannot detect fact-list
// changes; this test pins them.
func TestMergedRulesUsePolicyFactSets(t *testing.T) {
	p := YA2025Draft()
	want := map[string][]string{
		"MEDICAL_DENTAL":        {"no_other_claim", "qualifying_beneficiary", "qualifying_dental", "resident", "within_year"},
		"EDUCATION_SELF_SKILLS": {"no_other_claim", "own_education", "qualifying_course", "resident", "within_year"},
		"EDUCATION_SELF_OTHER":  {"no_other_claim", "own_education", "qualifying_course", "resident", "within_year"},
	}
	for id, facts := range want {
		got := append([]string(nil), findRule(t, p, id).RequiredFacts...)
		sort.Strings(got)
		if !reflect.DeepEqual(got, facts) {
			t.Fatalf("%s required facts = %v, want %v", id, got, facts)
		}
	}
}
