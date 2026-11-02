package eligibility

import (
	"sort"
	"strings"
	"testing"

	"github.com/fikrimastor/mytax/services/api/internal/policy"
)

// The eligibility engine (served by /v1/eligibility/preview) and the policy draft
// catalog (served by /v1/policy/catalog) are separate code paths for the same relief
// rules. If they disagree on required facts, the same claim can pass one screen and
// fail the other. Any divergence for a shared rule ID must fail this test.
func TestSharedRuleFactParityWithPolicyCatalog(t *testing.T) {
	shared := []string{
		"FIRST_HOME_LOW",
		"FIRST_HOME_MID",
		"LIFESTYLE_BOOKS",
		"LIFESTYLE_COURSE",
		"LIFESTYLE_DEVICE",
		"LIFESTYLE_INTERNET",
		"MEDICAL_VACCINATION",
	}
	engineFacts := map[string][]string{}
	for _, r := range Draft2025().Rules {
		engineFacts[r.ID] = r.RequiredFacts
	}
	policyFacts := map[string][]string{}
	for _, r := range policy.YA2025Draft().Rules {
		policyFacts[r.ID] = r.RequiredFacts
	}
	for _, id := range shared {
		ef, ok := engineFacts[id]
		if !ok {
			t.Fatalf("%s: missing from the eligibility catalog", id)
		}
		pf, ok := policyFacts[id]
		if !ok {
			t.Fatalf("%s: missing from the policy catalog", id)
		}
		got, want := factSetKey(ef), factSetKey(pf)
		if got != want {
			t.Fatalf("%s: required facts drifted\n eligibility: %s\n policy:      %s", id, got, want)
		}
	}
}

func factSetKey(facts []string) string {
	cp := append([]string(nil), facts...)
	sort.Strings(cp)
	return strings.Join(cp, "|")
}
