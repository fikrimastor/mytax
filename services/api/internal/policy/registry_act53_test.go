package policy

import (
	"strings"
	"testing"
)

// Entries that could not be confirmed against Act 53 must say so in their note, and
// their amounts must stay the LHDN-table values until a human resolves the difference.
func TestRegistryFlagsAct53Differences(t *testing.T) {
	want := map[string]int64{
		"DISABLED_SELF": 700000, "MEDICAL_LEARNING_DISABILITY": 600000, "CHILDCARE": 300000,
		"SSPN": 800000, "EDU_MED_INSURANCE": 400000, "DISABLED_SPOUSE": 600000, "DISABLED_CHILD": 800000,
	}
	byID := map[string]RegistryEntry{}
	for _, e := range ReliefRegistry2025() {
		byID[e.ID] = e
	}
	for id, sen := range want {
		e, ok := byID[id]
		if !ok {
			t.Fatalf("%s missing", id)
		}
		if e.MaxSen != sen {
			t.Fatalf("%s amount changed to %d without resolving the Act 53 difference", id, e.MaxSen)
		}
		if !strings.Contains(e.Note, "ACT53_DIFFERS") || !strings.Contains(e.Note, "Unresolved") {
			t.Fatalf("%s must carry an unresolved Act 53 flag: %q", id, e.Note)
		}
		if e.Status != "DRAFT_REVIEW_REQUIRED" {
			t.Fatalf("%s status %q", id, e.Status)
		}
	}
	for _, id := range []string{"EV_COMPOST", "MEDICAL_SCREENING", "FIRST_HOME_INTEREST_LOW", "FIRST_HOME_INTEREST_MID"} {
		if !strings.Contains(byID[id].Note, "ACT53_DIFFERS") {
			t.Fatalf("%s must carry an Act 53 flag", id)
		}
	}
	// Entries confirmed against the Act text must not carry the flag by accident.
	for _, id := range []string{"SELF", "PARENTS", "SELF_EDUCATION", "LIFESTYLE", "SPORTS", "SOCSO", "SPOUSE", "CHILD_UNDER_18", "CHILD_STUDY_HIGHER", "PRS"} {
		if strings.Contains(byID[id].Note, "ACT53_DIFFERS") {
			t.Fatalf("%s unexpectedly flagged", id)
		}
	}
}
