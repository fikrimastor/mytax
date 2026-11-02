package policy

import (
	"strings"
	"testing"
)

func registryByID() map[string]RegistryEntry {
	m := map[string]RegistryEntry{}
	for _, e := range ReliefRegistry2025() {
		m[e.ID] = e
	}
	return m
}

// Amounts stay the LHDN YA2025 table values, and every entry stays a draft.
func TestRegistryAct53FindingsKeepAmountsAndDraftStatus(t *testing.T) {
	want := map[string]int64{
		"DISABLED_SELF": 700000, "MEDICAL_LEARNING_DISABILITY": 600000, "CHILDCARE": 300000,
		"SSPN": 800000, "EDU_MED_INSURANCE": 400000, "DISABLED_SPOUSE": 600000, "DISABLED_CHILD": 800000,
		"FIRST_HOME_INTEREST_LOW": 700000, "FIRST_HOME_INTEREST_MID": 500000, "EV_COMPOST": 250000,
	}
	reg := registryByID()
	for id, sen := range want {
		e, ok := reg[id]
		if !ok || e.MaxSen != sen || e.Status != "DRAFT_REVIEW_REQUIRED" {
			t.Fatalf("%s changed or approved: %+v", id, e)
		}
	}
}

func TestRegistryAct53FindingKinds(t *testing.T) {
	reg := registryByID()
	kinds := map[string][]string{
		Act53Confirmed:   {"SSPN", "EV_COMPOST", "DISABLED_SPOUSE", "DISABLED_CHILD", "FIRST_HOME_INTEREST_LOW", "FIRST_HOME_INTEREST_MID"},
		Act53YAUncertain: {"DISABLED_SELF", "CHILDCARE", "EDU_MED_INSURANCE"},
		Act53Differs:     {"MEDICAL_LEARNING_DISABILITY"},
		Act53Scope:       {"MEDICAL_SCREENING", "PARENTS"},
	}
	for kind, ids := range kinds {
		for _, id := range ids {
			if !strings.Contains(reg[id].Note, " | "+kind+": ") {
				t.Fatalf("%s should carry %s: %q", id, kind, reg[id].Note)
			}
		}
	}
	// Only the real difference may say "Unresolved".
	for id, e := range reg {
		if strings.Contains(e.Note, "Unresolved") && id != "MEDICAL_LEARNING_DISABILITY" {
			t.Fatalf("%s says Unresolved without being a difference", id)
		}
	}
	// Uncertain entries must say the text does not settle the year.
	for _, id := range kinds[Act53YAUncertain] {
		if !strings.Contains(reg[id].Note, "not") {
			t.Fatalf("%s must explain the year uncertainty", id)
		}
	}
	// Entries confirmed only by the amount must not carry a finding by accident.
	for _, id := range []string{"SELF", "SELF_EDUCATION", "LIFESTYLE", "SPORTS", "SOCSO", "SPOUSE", "CHILD_UNDER_18", "CHILD_STUDY_HIGHER", "PRS"} {
		if strings.Contains(reg[id].Note, "ACT53_") {
			t.Fatalf("%s unexpectedly flagged", id)
		}
	}
}
