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
		"MEDICAL_SCREENING": 100000, "PARENTS": 800000, "MEDICAL_VACCINATION": 100000,
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
		Act53Consistent: {"SSPN", "EV_COMPOST", "DISABLED_SELF", "DISABLED_SPOUSE", "DISABLED_CHILD", "EDU_MED_INSURANCE",
			"MEDICAL_LEARNING_DISABILITY", "CHILDCARE", "FIRST_HOME_INTEREST_LOW", "FIRST_HOME_INTEREST_MID", "MEDICAL_SCREENING", "PARENTS"},
		Act53Scope: {"MEDICAL_VACCINATION"},
	}
	for kind, ids := range kinds {
		for _, id := range ids {
			if !strings.Contains(reg[id].Note, " | "+kind+": ") {
				t.Fatalf("%s should carry %s: %q", id, kind, reg[id].Note)
			}
		}
	}
	// Consistent entries must cite the amending Finance Act (or the Act section for the first-home rule).
	for _, id := range kinds[Act53Consistent] {
		if !strings.Contains(reg[id].Note, "Finance Act 2024") {
			t.Fatalf("%s must cite Finance Act 2024: %q", id, reg[id].Note)
		}
	}
	// No entry is currently unresolved or year-uncertain.
	for id, e := range reg {
		if strings.Contains(e.Note, Act53Differs) || strings.Contains(e.Note, Act53YAUncertain) || strings.Contains(e.Note, "Unresolved") {
			t.Fatalf("%s unexpectedly unresolved: %q", id, e.Note)
		}
	}
	// Entries with no amending-law finding must not carry a finding by accident.
	for _, id := range []string{"SELF", "SELF_EDUCATION", "LIFESTYLE", "SPORTS", "SOCSO", "SPOUSE", "CHILD_UNDER_18", "CHILD_STUDY_HIGHER", "PRS"} {
		if strings.Contains(reg[id].Note, "ACT53_") {
			t.Fatalf("%s unexpectedly flagged", id)
		}
	}
}

// YA2026 changes must not leak into the YA2025 values.
func TestRegistryDoesNotUseYA2026Values(t *testing.T) {
	reg := registryByID()
	if reg["MEDICAL_LEARNING_DISABILITY"].MaxSen != 600000 {
		t.Fatal("learning disability must stay RM6,000 for YA2025 (RM10,000 starts YA2026)")
	}
	for id := range reg {
		if strings.Contains(id, "TOURIST") || strings.Contains(id, "CCTV") || strings.Contains(id, "GRINDER") {
			t.Fatalf("YA2026-only relief %s must not be in the YA2025 registry", id)
		}
	}
}
