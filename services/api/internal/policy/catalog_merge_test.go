package policy

import "testing"

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
