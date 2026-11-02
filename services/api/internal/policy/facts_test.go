package policy

import (
	"regexp"
	"sort"
	"testing"
)

var factName = regexp.MustCompile(`^[a-z0-9_]+$`)

func TestCanonicalFactsAreWellFormed(t *testing.T) {
	if !sort.StringsAreSorted(CanonicalFacts) {
		t.Fatal("CanonicalFacts must stay sorted")
	}
	seen := map[string]bool{}
	for _, f := range CanonicalFacts {
		if !factName.MatchString(f) {
			t.Fatalf("fact %q is not snake_case", f)
		}
		if seen[f] {
			t.Fatalf("fact %q listed twice", f)
		}
		seen[f] = true
	}
}

func TestEveryRuleFactIsCanonicalAndEveryCanonicalFactIsUsed(t *testing.T) {
	used := map[string]bool{}
	for _, r := range YA2025Draft().Rules {
		for _, f := range r.RequiredFacts {
			if !IsCanonicalFact(f) {
				t.Fatalf("%s requires %q, which is not in CanonicalFacts", r.ID, f)
			}
			used[f] = true
		}
	}
	for _, f := range CanonicalFacts {
		if !used[f] {
			t.Fatalf("canonical fact %q is used by no rule; remove it or add the rule", f)
		}
	}
}
