# Unify the Relief Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `internal/policy` the only relief rule catalog and the only claim `Evaluate`, delete `internal/eligibility`, and keep every safeguard the eligibility engine has today.

**Architecture:** First record what `eligibility.Evaluate` returns today for a matrix of claims (golden data). Then extend `policy` (rule fields, canonical fact list, one `Evaluate` with exclusive families, expense keys, overflow and claim-count guards), prove the new engine reproduces the recorded behaviour, switch `/v1/eligibility/preview` to it, and delete `eligibility`.

**Tech Stack:** Go 1.23 module `github.com/fikrimastor/mytax/services/api`, Echo, standard `testing`. All commands run from `services/api` unless stated. Spec: `docs/superpowers/specs/2026-11-03-unify-relief-engine-design.md`.

**Rules for every task (from `AGENTS.md`):** fail closed; nothing becomes Approved; money is `int64` sen; no real taxpayer data; tests that skip are not proof. Run `gofmt -l` on files you touch and fix them. Commit only the files named in the task.

---

## File structure

| Path | Action | Responsibility |
|---|---|---|
| `services/api/internal/policy/testdata/relief_cases.json` | create | Claim matrix (inputs) plus outputs recorded from the old eligibility engine |
| `services/api/internal/eligibility/golden_record_test.go` | create (deleted in Task 5) | Generates and checks the recorded outputs against the old engine |
| `services/api/internal/policy/engine.go` | rewrite | Types and the single `Evaluate` |
| `services/api/internal/policy/catalog.go` | modify | Add `MEDICAL_DENTAL`, `ExclusiveFamily` on first-home rules |
| `services/api/internal/policy/facts.go` | create | Canonical fact list |
| `services/api/internal/policy/facts_test.go` | create | Every rule fact must be canonical |
| `services/api/internal/policy/catalog_merge_test.go` | create | Tests for the merged rules |
| `services/api/internal/policy/engine_unified_test.go` | create | New engine behaviour tests |
| `services/api/internal/policy/equivalence_test.go` | create | Replays the recorded matrix through `policy.Evaluate` |
| `services/api/internal/policy/firsthome_cap_test.go` | create | Ported first-home tests |
| `services/api/internal/policy/engine_test.go`, `lifestyle_test.go` | modify | Move to the new claim shape |
| `services/api/internal/assessment/assessment.go`, `assessment_test.go` | modify | New claim shape, drop duplicate guards |
| `services/api/cmd/server/main.go` | modify | Serve `/v1/eligibility/preview` from policy |
| `services/api/internal/eligibility/` | delete | Replaced by policy |
| `docs/*`, `AGENTS.md` | modify | Keep docs true |

---

### Task 0: Isolated workspace

**Files:** none.

- [ ] **Step 1: Create a worktree on a new branch**

Run from the main checkout `/Users/fikrimastor/Sites/fikrimastor/mytax`:

```bash
git fetch -q origin main
git worktree add -b refactor/unify-relief-engine ../mytax-unify docs/unify-relief-engine
cd ../mytax-unify/services/api && git log --oneline -3 && go test ./... 2>&1 | tail -5
```

Expected: the branch holds the spec commit on top of `origin/main`; `go test ./...` prints `ok` for every package (persistence tests report skips without a database; that is expected here).

---

### Task 1: Record today's eligibility behaviour (characterization)

**Files:**
- Create: `services/api/internal/policy/testdata/relief_cases.json`
- Create: `services/api/internal/eligibility/golden_record_test.go`

- [ ] **Step 1: Write the claim matrix (inputs only)**

Create `services/api/internal/policy/testdata/relief_cases.json`. `expect` is left empty; Step 3 fills it from the old engine. `missing_facts` removes a fact the rule requires; `false_facts` sets a required fact to `false`; `repeat` makes N copies of the first claim with ids `c0000…` and keys `e0000…`.

```json
[
  {"name": "lifestyle exactly at cap", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 250000, "evidence_reference": "inv"}]},
  {"name": "lifestyle cap plus one sen", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 250001, "evidence_reference": "inv"}]},
  {"name": "lifestyle cap shared by four categories", "claims": [
    {"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 60000, "evidence_reference": "inv"},
    {"id": "b", "expense_key": "e2", "rule_id": "LIFESTYLE_INTERNET", "amount_sen": 90000, "evidence_reference": "inv"},
    {"id": "c", "expense_key": "e3", "rule_id": "LIFESTYLE_DEVICE", "amount_sen": 80000, "evidence_reference": "inv"},
    {"id": "d", "expense_key": "e4", "rule_id": "LIFESTYLE_COURSE", "amount_sen": 70000, "evidence_reference": "inv"}]},
  {"name": "group cap with claim ids in reverse order", "claims": [
    {"id": "b", "expense_key": "e2", "rule_id": "LIFESTYLE_INTERNET", "amount_sen": 200000, "evidence_reference": "inv"},
    {"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 190000, "evidence_reference": "inv"}]},
  {"name": "same expense key on two rules", "claims": [
    {"id": "a", "expense_key": "same", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 50000, "evidence_reference": "inv"},
    {"id": "b", "expense_key": "same", "rule_id": "LIFESTYLE_INTERNET", "amount_sen": 50000, "evidence_reference": "inv"}]},
  {"name": "missing expense key", "claims": [{"id": "a", "expense_key": "", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 100000, "evidence_reference": "inv"}]},
  {"name": "missing evidence reference", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 100000, "evidence_reference": ""}]},
  {"name": "unknown rule", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_UNKNOWN", "amount_sen": 1, "evidence_reference": "inv"}]},
  {"name": "vaccination without qualifying vaccine", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 200000, "evidence_reference": "inv", "missing_facts": ["qualifying_vaccine"]}]},
  {"name": "device used for business", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_DEVICE", "amount_sen": 100000, "evidence_reference": "inv", "false_facts": ["not_business_use"]}]},
  {"name": "blocked claim does not consume the cap", "claims": [
    {"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 100000, "evidence_reference": "inv", "missing_facts": ["qualifying_reading_material"]},
    {"id": "b", "expense_key": "e2", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 250000, "evidence_reference": "inv"}]},
  {"name": "first home tiers are mutually exclusive", "claims": [
    {"id": "a", "expense_key": "loan1", "rule_id": "FIRST_HOME_LOW", "amount_sen": 300000, "evidence_reference": "inv"},
    {"id": "b", "expense_key": "loan2", "rule_id": "FIRST_HOME_MID", "amount_sen": 300000, "evidence_reference": "inv"}]},
  {"name": "first home low tier cap plus one", "claims": [{"id": "a", "expense_key": "loan1", "rule_id": "FIRST_HOME_LOW", "amount_sen": 700001, "evidence_reference": "inv"}]},
  {"name": "first home mid tier cap plus one", "claims": [{"id": "a", "expense_key": "loan1", "rule_id": "FIRST_HOME_MID", "amount_sen": 500001, "evidence_reference": "inv"}]},
  {"name": "first home without co-claimant confirmation", "claims": [{"id": "a", "expense_key": "loan1", "rule_id": "FIRST_HOME_LOW", "amount_sen": 300000, "evidence_reference": "inv", "missing_facts": ["co_claimant_none"]}]},
  {"name": "lifestyle medical and education pools are separate", "claims": [
    {"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 250000, "evidence_reference": "inv"},
    {"id": "b", "expense_key": "e2", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 100000, "evidence_reference": "inv"},
    {"id": "c", "expense_key": "e3", "rule_id": "EDUCATION_UPSKILL", "amount_sen": 200000, "evidence_reference": "inv"}]},
  {"name": "vaccination over its own limit", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 200000, "evidence_reference": "inv"}]},
  {"name": "dental within limit", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "MEDICAL_DENTAL", "amount_sen": 50000, "evidence_reference": "inv"}]},
  {"name": "one course fee as education and lifestyle", "claims": [
    {"id": "a", "expense_key": "course9", "rule_id": "EDUCATION_UPSKILL", "amount_sen": 100000, "evidence_reference": "inv"},
    {"id": "b", "expense_key": "course9", "rule_id": "LIFESTYLE_COURSE", "amount_sen": 100000, "evidence_reference": "inv"}]},
  {"name": "duplicate claim id is rejected", "claims": [
    {"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 1, "evidence_reference": "inv"},
    {"id": "a", "expense_key": "e2", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 1, "evidence_reference": "inv"}]},
  {"name": "zero amount is rejected", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 0, "evidence_reference": "inv"}]},
  {"name": "negative amount is rejected", "claims": [{"id": "a", "expense_key": "e1", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": -1, "evidence_reference": "inv"}]},
  {"name": "exactly 1000 claims are accepted", "repeat": 1000, "claims": [{"id": "x", "expense_key": "x", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 10000, "evidence_reference": "inv"}]},
  {"name": "1001 claims are rejected", "repeat": 1001, "claims": [{"id": "x", "expense_key": "x", "rule_id": "LIFESTYLE_BOOKS", "amount_sen": 10000, "evidence_reference": "inv"}]}
]
```

- [ ] **Step 2: Write the recorder**

Create `services/api/internal/eligibility/golden_record_test.go`:

```go
package eligibility

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// This file freezes what the eligibility engine returns today so that the unified
// policy engine can be proven equivalent. It is deleted together with this package;
// the JSON file it writes stays in internal/policy/testdata.

var update = flag.Bool("update", false, "rewrite the expect blocks in relief_cases.json")

const casesPath = "../policy/testdata/relief_cases.json"

type caseClaim struct {
	ID                string   `json:"id"`
	ExpenseKey        string   `json:"expense_key"`
	RuleID            string   `json:"rule_id"`
	AmountSen         int64    `json:"amount_sen"`
	EvidenceReference string   `json:"evidence_reference"`
	MissingFacts      []string `json:"missing_facts,omitempty"`
	FalseFacts        []string `json:"false_facts,omitempty"`
}

type caseDecision struct {
	ID           string   `json:"id"`
	CandidateSen int64    `json:"candidate_sen"`
	Reasons      []string `json:"reasons"`
}

type caseExpect struct {
	Error               bool           `json:"error"`
	DecisionCount       int            `json:"decision_count,omitempty"`
	Decisions           []caseDecision `json:"decisions,omitempty"`
	ProvisionalTotalSen int64          `json:"provisional_total_sen"`
}

type reliefCase struct {
	Name   string      `json:"name"`
	Repeat int         `json:"repeat,omitempty"`
	Claims []caseClaim `json:"claims"`
	Expect caseExpect  `json:"expect"`
}

func buildClaims(cat Catalog, c reliefCase) []Claim {
	templates := c.Claims
	if c.Repeat > 0 {
		templates = nil
		for i := 0; i < c.Repeat; i++ {
			t := c.Claims[0]
			t.ID = fmt.Sprintf("c%04d", i)
			t.ExpenseKey = fmt.Sprintf("e%04d", i)
			templates = append(templates, t)
		}
	}
	required := map[string][]string{}
	for _, r := range cat.Rules {
		required[r.ID] = r.RequiredFacts
	}
	out := make([]Claim, 0, len(templates))
	for _, t := range templates {
		facts := map[string]bool{}
		for _, f := range required[t.RuleID] {
			facts[f] = true
		}
		for _, f := range t.MissingFacts {
			delete(facts, f)
		}
		for _, f := range t.FalseFacts {
			facts[f] = false
		}
		out = append(out, Claim{ID: t.ID, ExpenseKey: t.ExpenseKey, RuleID: t.RuleID, AmountSen: t.AmountSen, EvidenceReference: t.EvidenceReference, Facts: facts})
	}
	return out
}

func record(c reliefCase) caseExpect {
	cat := Draft2025()
	res, err := Evaluate(cat, buildClaims(cat, c))
	if err != nil {
		return caseExpect{Error: true}
	}
	exp := caseExpect{ProvisionalTotalSen: res.ProvisionalTotalSen}
	if c.Repeat > 0 {
		exp.DecisionCount = len(res.Decisions)
		return exp
	}
	for _, d := range res.Decisions {
		exp.Decisions = append(exp.Decisions, caseDecision{ID: d.ID, CandidateSen: d.CandidateSen, Reasons: d.Reasons})
	}
	return exp
}

func TestRecordedEligibilityBehaviour(t *testing.T) {
	raw, err := os.ReadFile(casesPath)
	if err != nil {
		t.Fatal(err)
	}
	var cases []reliefCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for i := range cases {
		got := record(cases[i])
		if *update {
			cases[i].Expect = got
			continue
		}
		if !reflect.DeepEqual(got, cases[i].Expect) {
			t.Errorf("%s:\n got  %+v\n want %+v", cases[i].Name, got, cases[i].Expect)
		}
	}
	if *update {
		out, err := json.MarshalIndent(cases, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(casesPath, append(out, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
```

- [ ] **Step 3: Record the outputs, then check them**

```bash
go test ./internal/eligibility -run TestRecordedEligibilityBehaviour -update
go test ./internal/eligibility -run TestRecordedEligibilityBehaviour -count=1 -v
```

Expected: first command `ok`; second `--- PASS: TestRecordedEligibilityBehaviour`. Then spot-check the file: `grep -c '"error": true' internal/policy/testdata/relief_cases.json` prints `4` (duplicate id, zero, negative, 1001 claims), and the case "first home tiers are mutually exclusive" shows `"MUTUALLY_EXCLUSIVE_RELIEF"` on claim `b`. If those counts differ, stop and report.

- [ ] **Step 4: Commit**

```bash
gofmt -l internal/eligibility/golden_record_test.go
git add internal/eligibility/golden_record_test.go internal/policy/testdata/relief_cases.json
git commit -m "test(eligibility): record current engine behaviour as golden relief cases"
```

---

### Task 2: Merge the eligibility-only rules into the policy catalog

**Files:**
- Modify: `services/api/internal/policy/engine.go` (add one field to `Rule`)
- Modify: `services/api/internal/policy/catalog.go`
- Create: `services/api/internal/policy/catalog_merge_test.go`

- [ ] **Step 1: Write the failing test**

Create `services/api/internal/policy/catalog_merge_test.go`:

```go
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
```

- [ ] **Step 2: Run it to see it fail**

Run: `go test ./internal/policy -run 'TestFirstHomeTiers|TestDentalRule' -count=1`
Expected: compile error `r.ExclusiveFamily undefined` (build failure counts as failing).

- [ ] **Step 3: Add the field and the rules**

In `services/api/internal/policy/engine.go`, change the `Rule` struct to:

```go
type Rule struct {
 ID string `json:"id"`
 Group string `json:"group"`
 ExclusiveFamily string `json:"exclusive_family,omitempty"`
 LimitSen int64 `json:"limit_sen"`
 RequiredFacts []string `json:"required_facts"`
 Sources []Source `json:"sources"`
}
```

In `services/api/internal/policy/catalog.go`, add this line after the `MEDICAL_VACCINATION` rule:

```go
			{ID: "MEDICAL_DENTAL", Group: "MEDICAL_SELF_FAMILY", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_dental", "qualifying_beneficiary", "no_other_claim"}, Sources: src},
```

and add `ExclusiveFamily: "FIRST_HOME", ` right after `ID: "FIRST_HOME_LOW", ` and after `ID: "FIRST_HOME_MID", `. Also replace the comment above the first-home rules with:

```go
			// s.46B first-home interest. The two price tiers are mutually exclusive (ExclusiveFamily).
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/policy ./internal/eligibility -count=1`
Expected: `ok` for both (the eligibility parity test still passes because dental is not in its shared list).

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/policy/catalog_merge_test.go
git add internal/policy/engine.go internal/policy/catalog.go internal/policy/catalog_merge_test.go
git commit -m "feat(policy): add MEDICAL_DENTAL and the first-home exclusive family to the catalog"
```

---

### Task 3: Canonical fact list

**Files:**
- Create: `services/api/internal/policy/facts.go`
- Create: `services/api/internal/policy/facts_test.go`

- [ ] **Step 1: Write the failing test**

Create `services/api/internal/policy/facts_test.go`:

```go
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
```

- [ ] **Step 2: Run it to see it fail**

Run: `go test ./internal/policy -run 'TestCanonicalFacts|TestEveryRuleFact' -count=1`
Expected: compile error `undefined: CanonicalFacts`.

- [ ] **Step 3: Write the list**

Create `services/api/internal/policy/facts.go`:

```go
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
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/policy -count=1`
Expected: `ok`. If a fact is reported as unused or missing, fix the list (do not change rules).

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/policy/facts.go internal/policy/facts_test.go
git add internal/policy/facts.go internal/policy/facts_test.go
git commit -m "feat(policy): add the canonical fact list and a test that every rule uses it"
```

---

### Task 4: One engine in policy

This task changes the claim shape, so the policy tests and `assessment` change in the same commit to keep the build green.

**Files:**
- Rewrite: `services/api/internal/policy/engine.go`
- Modify: `services/api/internal/policy/engine_test.go`, `services/api/internal/policy/lifestyle_test.go`
- Modify: `services/api/internal/assessment/assessment.go`, `services/api/internal/assessment/assessment_test.go`
- Create: `services/api/internal/policy/engine_unified_test.go`, `services/api/internal/policy/equivalence_test.go`

- [ ] **Step 1: Write the equivalence test (it will not compile yet)**

Create `services/api/internal/policy/equivalence_test.go`:

```go
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Replays testdata/relief_cases.json, recorded from the retired eligibility engine,
// through policy.Evaluate. The only allowed differences are the renames below.

type recordedClaim struct {
	ID                string   `json:"id"`
	ExpenseKey        string   `json:"expense_key"`
	RuleID            string   `json:"rule_id"`
	AmountSen         int64    `json:"amount_sen"`
	EvidenceReference string   `json:"evidence_reference"`
	MissingFacts      []string `json:"missing_facts"`
	FalseFacts        []string `json:"false_facts"`
}

type recordedDecision struct {
	ID           string   `json:"id"`
	CandidateSen int64    `json:"candidate_sen"`
	Reasons      []string `json:"reasons"`
}

type recordedCase struct {
	Name   string          `json:"name"`
	Repeat int             `json:"repeat"`
	Claims []recordedClaim `json:"claims"`
	Expect struct {
		Error               bool               `json:"error"`
		DecisionCount       int                `json:"decision_count"`
		Decisions           []recordedDecision `json:"decisions"`
		ProvisionalTotalSen int64              `json:"provisional_total_sen"`
	} `json:"expect"`
}

var ruleRename = map[string]string{
	"EDUCATION_UPSKILL": "EDUCATION_SELF_SKILLS",
	"EDUCATION_OTHER":   "EDUCATION_SELF_OTHER",
}

func renameReason(r string) string {
	switch {
	case r == "UNKNOWN_RULE":
		return "UNKNOWN_RELIEF_RULE"
	case r == "DRAFT_LEGAL_RULE_NOT_APPROVED":
		return "DRAFT_POLICY_NOT_APPROVED"
	case strings.HasPrefix(r, "FACT_NOT_VERIFIED:"):
		return "UNVERIFIED_FACT:" + strings.TrimPrefix(r, "FACT_NOT_VERIFIED:")
	}
	return r
}

func buildRecordedClaims(p Package, c recordedCase) []Claim {
	templates := c.Claims
	if c.Repeat > 0 {
		templates = nil
		for i := 0; i < c.Repeat; i++ {
			t := c.Claims[0]
			t.ID = fmt.Sprintf("c%04d", i)
			t.ExpenseKey = fmt.Sprintf("e%04d", i)
			templates = append(templates, t)
		}
	}
	required := map[string][]string{}
	for _, r := range p.Rules {
		required[r.ID] = r.RequiredFacts
	}
	out := make([]Claim, 0, len(templates))
	for _, t := range templates {
		rule := t.RuleID
		if n, ok := ruleRename[rule]; ok {
			rule = n
		}
		facts := map[string]bool{}
		for _, f := range required[rule] {
			facts[f] = true
		}
		for _, f := range t.MissingFacts {
			delete(facts, f)
		}
		for _, f := range t.FalseFacts {
			facts[f] = false
		}
		out = append(out, Claim{ID: t.ID, RuleID: rule, AmountSen: t.AmountSen, ExpenseKey: t.ExpenseKey, EvidenceReference: t.EvidenceReference, Facts: facts})
	}
	return out
}

func TestUnifiedEngineMatchesRecordedEligibilityBehaviour(t *testing.T) {
	raw, err := os.ReadFile("testdata/relief_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []recordedCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 20 {
		t.Fatalf("only %d recorded cases; the file looks truncated", len(cases))
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			p := YA2025Draft()
			res, err := Evaluate(p, buildRecordedClaims(p, c))
			if c.Expect.Error {
				if err == nil {
					t.Fatal("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if res.ProvisionalTotalSen != c.Expect.ProvisionalTotalSen {
				t.Fatalf("total %d, recorded %d", res.ProvisionalTotalSen, c.Expect.ProvisionalTotalSen)
			}
			if res.ApprovedTotalSen != 0 || res.FilingReady {
				t.Fatalf("draft leaked approval: %+v", res)
			}
			if c.Repeat > 0 {
				if len(res.Decisions) != c.Expect.DecisionCount {
					t.Fatalf("%d decisions, recorded %d", len(res.Decisions), c.Expect.DecisionCount)
				}
				return
			}
			if len(res.Decisions) != len(c.Expect.Decisions) {
				t.Fatalf("%d decisions, recorded %d", len(res.Decisions), len(c.Expect.Decisions))
			}
			for i, want := range c.Expect.Decisions {
				got := res.Decisions[i]
				if got.ClaimID != want.ID || got.ProvisionalAmountSen != want.CandidateSen {
					t.Fatalf("decision %d: got %s/%d, recorded %s/%d", i, got.ClaimID, got.ProvisionalAmountSen, want.ID, want.CandidateSen)
				}
				if got.Status != "REVIEW_REQUIRED" {
					t.Fatalf("decision %d status %q", i, got.Status)
				}
				if len(got.Reasons) != len(want.Reasons) {
					t.Fatalf("decision %d reasons %v, recorded %v", i, got.Reasons, want.Reasons)
				}
				for j, r := range want.Reasons {
					if got.Reasons[j] != renameReason(r) {
						t.Fatalf("decision %d reason %d: got %q, recorded %q (renamed %q)", i, j, got.Reasons[j], r, renameReason(r))
					}
				}
			}
		})
	}
}
```

- [ ] **Step 2: Replace `internal/policy/engine.go` with the unified engine**

Overwrite the whole file:

```go
// Package policy holds the draft relief rules and the single engine that evaluates
// claims against them. It does not interpret the law or certify tax entitlement.
package policy

import (
	"errors"
	"math"
	"sort"
)

const (
	supportedYear = 2025
	maxClaims     = 1000
)

type Status string

const (
	Draft    Status = "DRAFT"
	Approved Status = "APPROVED"
)

type Source struct {
	URL         string `json:"url"`
	Section     string `json:"section"`
	RetrievedOn string `json:"retrieved_on"`
}

type Rule struct {
	ID              string   `json:"id"`
	Group           string   `json:"group"`
	ExclusiveFamily string   `json:"exclusive_family,omitempty"`
	LimitSen        int64    `json:"limit_sen"`
	RequiredFacts   []string `json:"required_facts"`
	Sources         []Source `json:"sources"`
}

type Group struct {
	ID       string `json:"id"`
	LimitSen int64  `json:"limit_sen"`
}

type Package struct {
	AssessmentYear int     `json:"assessment_year"`
	Version        string  `json:"version"`
	Status         Status  `json:"status"`
	Rules          []Rule  `json:"rules"`
	Groups         []Group `json:"groups"`
}

// Claim is a candidate relief claim. The server derives duplicates from ExpenseKey and
// evidence from EvidenceReference; the caller cannot assert either as a boolean.
type Claim struct {
	ID                string          `json:"id"`
	RuleID            string          `json:"rule_id"`
	AmountSen         int64           `json:"amount_sen"`
	ExpenseKey        string          `json:"expense_key"`
	EvidenceReference string          `json:"evidence_reference"`
	Facts             map[string]bool `json:"facts"`
}

type Decision struct {
	ClaimID              string   `json:"claim_id"`
	Status               string   `json:"status"`
	ProvisionalAmountSen int64    `json:"provisional_amount_sen"`
	ApprovedAmountSen    int64    `json:"approved_amount_sen"`
	Reasons              []string `json:"reasons"`
}

type Result struct {
	AssessmentYear      int        `json:"assessment_year"`
	PackageVersion      string     `json:"package_version"`
	PackageStatus       Status     `json:"package_status"`
	Decisions           []Decision `json:"decisions"`
	ProvisionalTotalSen int64      `json:"provisional_total_sen"`
	ApprovedTotalSen    int64      `json:"approved_total_sen"`
	FilingReady         bool       `json:"filing_ready"`
	Notice              string     `json:"notice"`
}

// Evaluate screens claims against a package. Claims are allocated in claim-ID order so the
// result does not depend on request order. Anything unknown, unverified, duplicated,
// mutually exclusive or without evidence is REVIEW_REQUIRED with a zero amount.
func Evaluate(p Package, claims []Claim) (Result, error) {
	if p.AssessmentYear != supportedYear {
		return Result{}, errors.New("unsupported assessment year")
	}
	if p.Version == "" {
		return Result{}, errors.New("package version required")
	}
	if p.Status != Draft && p.Status != Approved {
		return Result{}, errors.New("unknown package status")
	}
	if len(claims) > maxClaims {
		return Result{}, errors.New("too many claims")
	}
	rules := map[string]Rule{}
	groups := map[string]Group{}
	for _, g := range p.Groups {
		if g.ID == "" || g.LimitSen < 0 {
			return Result{}, errors.New("invalid group")
		}
		if _, ok := groups[g.ID]; ok {
			return Result{}, errors.New("duplicate group")
		}
		groups[g.ID] = g
	}
	for _, rule := range p.Rules {
		if rule.ID == "" || rule.LimitSen < 0 {
			return Result{}, errors.New("invalid rule")
		}
		if _, ok := rules[rule.ID]; ok {
			return Result{}, errors.New("duplicate rule")
		}
		if rule.Group != "" {
			if _, ok := groups[rule.Group]; !ok {
				return Result{}, errors.New("unknown group")
			}
		}
		rules[rule.ID] = rule
	}

	ordered := append([]Claim(nil), claims...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	usedRules := map[string]int64{}
	usedGroups := map[string]int64{}
	seenID := map[string]bool{}
	seenExpense := map[string]bool{}
	families := map[string]bool{}
	res := Result{
		AssessmentYear: p.AssessmentYear,
		PackageVersion: p.Version,
		PackageStatus:  p.Status,
		Decisions:      make([]Decision, 0, len(ordered)),
		Notice:         "Draft rules are not legally approved and never authorize a tax claim.",
	}
	for _, c := range ordered {
		if c.ID == "" || c.AmountSen <= 0 {
			return Result{}, errors.New("claim ID and positive amount required")
		}
		if seenID[c.ID] {
			return Result{}, errors.New("duplicate claim ID")
		}
		seenID[c.ID] = true
		d := Decision{ClaimID: c.ID, Status: "REVIEW_REQUIRED", Reasons: []string{}}
		rule, ok := rules[c.RuleID]
		if !ok {
			d.Reasons = append(d.Reasons, "UNKNOWN_RELIEF_RULE")
		}
		if c.ExpenseKey == "" {
			d.Reasons = append(d.Reasons, "EXPENSE_IDENTIFIER_MISSING")
		} else if seenExpense[c.ExpenseKey] {
			d.Reasons = append(d.Reasons, "DUPLICATE_EXPENSE")
		}
		if c.ExpenseKey != "" {
			seenExpense[c.ExpenseKey] = true
		}
		if c.EvidenceReference == "" {
			d.Reasons = append(d.Reasons, "EVIDENCE_MISSING")
		}
		if ok {
			if rule.ExclusiveFamily != "" && families[rule.ExclusiveFamily] {
				d.Reasons = append(d.Reasons, "MUTUALLY_EXCLUSIVE_RELIEF")
			}
			for _, key := range rule.RequiredFacts {
				if !c.Facts[key] {
					d.Reasons = append(d.Reasons, "UNVERIFIED_FACT:"+key)
				}
			}
		}
		if len(d.Reasons) == 0 {
			available := rule.LimitSen - usedRules[rule.ID]
			if rule.Group != "" {
				if g := groups[rule.Group].LimitSen - usedGroups[rule.Group]; g < available {
					available = g
				}
			}
			if available < 0 {
				available = 0
			}
			allowed := c.AmountSen
			if allowed > available {
				allowed = available
			}
			if res.ProvisionalTotalSen > math.MaxInt64-allowed {
				return Result{}, errors.New("overflow")
			}
			d.ProvisionalAmountSen = allowed
			res.ProvisionalTotalSen += allowed
			usedRules[rule.ID] += allowed
			if rule.Group != "" {
				usedGroups[rule.Group] += allowed
			}
			if allowed > 0 && rule.ExclusiveFamily != "" {
				families[rule.ExclusiveFamily] = true
			}
			if p.Status == Approved {
				d.Status = "POLICY_APPROVED_PRELIMINARY"
				d.ApprovedAmountSen = allowed
				res.ApprovedTotalSen += allowed
			} else {
				d.Reasons = append(d.Reasons, "DRAFT_POLICY_NOT_APPROVED")
			}
			if allowed < c.AmountSen {
				d.Reasons = append(d.Reasons, "LIMIT_REACHED")
			}
		}
		res.Decisions = append(res.Decisions, d)
	}
	return res, nil
}
```

- [ ] **Step 3: Move the existing tests to the new claim shape**

`services/api/internal/policy/engine_test.go` (replace the file):

```go
package policy

import "testing"

func facts(keys ...string) map[string]bool {
	m := map[string]bool{}
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func validClaim(id, rule string, amt int64) Claim {
	keys := []string{"resident", "within_year", "internet_in_own_name", "qualifying_expense", "no_other_claim", "qualifying_reading_material", "qualifying_beneficiary"}
	return Claim{ID: id, RuleID: rule, AmountSen: amt, ExpenseKey: "exp-" + id, EvidenceReference: "invoice", Facts: facts(keys...)}
}

func TestDraftNeverApproves(t *testing.T) {
	p := YA2025Draft()
	r, err := Evaluate(p, []Claim{validClaim("a", "LIFESTYLE_INTERNET", 200000)})
	if err != nil || r.ApprovedTotalSen != 0 || r.Decisions[0].ApprovedAmountSen != 0 || r.Decisions[0].ProvisionalAmountSen != 200000 {
		t.Fatalf("draft leaked approval %+v %v", r, err)
	}
}

func TestSharedCapAndDeterminism(t *testing.T) {
	p := YA2025Draft()
	p.Status = Approved // arithmetic-only test fixture; does NOT signify tax-policy sign-off
	claims := []Claim{validClaim("b", "LIFESTYLE_INTERNET", 200000), validClaim("a", "LIFESTYLE_BOOKS", 180000)}
	a, err := Evaluate(p, claims)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Evaluate(p, []Claim{claims[1], claims[0]})
	if err != nil {
		t.Fatal(err)
	}
	if a.ApprovedTotalSen != 250000 || b.ApprovedTotalSen != 250000 || a.Decisions[0].ApprovedAmountSen != 180000 || a.Decisions[1].ApprovedAmountSen != 70000 {
		t.Fatalf("cap allocation wrong %+v %+v", a, b)
	}
}

func TestMissingEvidenceAndFacts(t *testing.T) {
	p := YA2025Draft()
	c := validClaim("a", "LIFESTYLE_INTERNET", 50000)
	c.EvidenceReference = ""
	r, _ := Evaluate(p, []Claim{c})
	if r.Decisions[0].ProvisionalAmountSen != 0 {
		t.Fatal("evidence missing must block")
	}
	c.EvidenceReference = "invoice"
	delete(c.Facts, "internet_in_own_name")
	r, _ = Evaluate(p, []Claim{c})
	if r.Decisions[0].ProvisionalAmountSen != 0 {
		t.Fatal("required fact missing must block")
	}
}

func TestRejectedInputs(t *testing.T) {
	p := YA2025Draft()
	if _, err := Evaluate(p, []Claim{{ID: "x", RuleID: "LIFESTYLE_INTERNET", AmountSen: -1}}); err == nil {
		t.Fatal("negative amount accepted")
	}
	if _, err := Evaluate(p, []Claim{validClaim("x", "LIFESTYLE_INTERNET", 1), validClaim("x", "LIFESTYLE_INTERNET", 1)}); err == nil {
		t.Fatal("duplicate ID accepted")
	}
}
```

`services/api/internal/policy/lifestyle_test.go` — make these three edits:

1. In `claimFor`, replace the return line with:
```go
	return Claim{ID: id, RuleID: rule, AmountSen: amt, ExpenseKey: "exp-" + id, EvidenceReference: "invoice", Facts: f}
```
2. In `TestLifestyleEvidenceDuplicateAndUnknownRuleFailClosed`, replace `c.EvidencePresent = false` with `c.EvidenceReference = ""`; replace the block
```go
		c = claimFor(t, p, "a", id, 100000)
		c.Duplicate = true
		r, _ = Evaluate(p, []Claim{c})
		if r.ApprovedTotalSen != 0 || r.Decisions[0].Reasons[0] != "DUPLICATE_EXPENSE" {
			t.Fatalf("%s: duplicate must block", id)
		}
```
with
```go
		first := claimFor(t, p, "a", id, 100000)
		second := claimFor(t, p, "b", id, 100000)
		second.ExpenseKey = first.ExpenseKey
		r, _ = Evaluate(p, []Claim{first, second})
		if r.Decisions[1].ApprovedAmountSen != 0 || r.Decisions[1].Reasons[0] != "DUPLICATE_EXPENSE" {
			t.Fatalf("%s: repeated expense key must block the second claim: %+v", id, r.Decisions[1])
		}
```
and replace the unknown-rule claim literal `{ID: "a", RuleID: "LIFESTYLE_UNKNOWN", AmountSen: 1, EvidencePresent: true}` with `{ID: "a", RuleID: "LIFESTYLE_UNKNOWN", AmountSen: 1, ExpenseKey: "k", EvidenceReference: "x"}`.

`services/api/internal/assessment/assessment.go`: delete the line `if len(req.Claims)>1000{return Response{},errors.New("too many claims")}`, and replace the block from `var excluded int64` through its closing loop brace with `excluded:=screened.ProvisionalTotalSen`. The `math` import stays (still used by the income check).

`services/api/internal/assessment/assessment_test.go`: in `TestDraftClaimsNeverReduceTax` replace `EvidencePresent:true,` with `ExpenseKey:"inv-1",EvidenceReference:"invoice",`.

- [ ] **Step 4: Write the new engine tests**

Create `services/api/internal/policy/engine_unified_test.go`:

```go
package policy

import (
	"fmt"
	"math"
	"testing"
)

func TestExclusiveFamilyBlocksTheSecondFirstHomeTier(t *testing.T) {
	p := YA2025Draft()
	low := claimFor(t, p, "a", "FIRST_HOME_LOW", 300000)
	mid := claimFor(t, p, "b", "FIRST_HOME_MID", 300000)
	r, err := Evaluate(p, []Claim{low, mid})
	if err != nil || r.ProvisionalTotalSen != 300000 {
		t.Fatalf("%+v %v", r, err)
	}
	if r.Decisions[1].ProvisionalAmountSen != 0 || r.Decisions[1].Reasons[0] != "MUTUALLY_EXCLUSIVE_RELIEF" {
		t.Fatalf("second tier must be blocked: %+v", r.Decisions[1])
	}
}

func TestExpenseKeyIsRequiredAndCannotBeReused(t *testing.T) {
	p := YA2025Draft()
	none := claimFor(t, p, "a", "LIFESTYLE_BOOKS", 100)
	none.ExpenseKey = ""
	r, _ := Evaluate(p, []Claim{none})
	if r.Decisions[0].Reasons[0] != "EXPENSE_IDENTIFIER_MISSING" || r.ProvisionalTotalSen != 0 {
		t.Fatalf("missing key must block: %+v", r.Decisions[0])
	}
	a := claimFor(t, p, "a", "LIFESTYLE_BOOKS", 100)
	b := claimFor(t, p, "b", "LIFESTYLE_INTERNET", 100)
	b.ExpenseKey = a.ExpenseKey
	r, _ = Evaluate(p, []Claim{a, b})
	if r.ProvisionalTotalSen != 100 || r.Decisions[1].Reasons[0] != "DUPLICATE_EXPENSE" {
		t.Fatalf("reused key must block the second claim: %+v", r)
	}
}

func TestClaimCountLimit(t *testing.T) {
	p := YA2025Draft()
	mk := func(n int) []Claim {
		out := make([]Claim, 0, n)
		for i := 0; i < n; i++ {
			c := claimFor(t, p, fmt.Sprintf("c%04d", i), "LIFESTYLE_BOOKS", 1)
			out = append(out, c)
		}
		return out
	}
	if _, err := Evaluate(p, mk(1000)); err != nil {
		t.Fatalf("1000 claims must be accepted: %v", err)
	}
	if _, err := Evaluate(p, mk(1001)); err == nil {
		t.Fatal("1001 claims accepted")
	}
}

func TestOnlyYear2025IsEvaluated(t *testing.T) {
	for _, year := range []int{2023, 2024, 2026} {
		p := YA2025Draft()
		p.AssessmentYear = year
		if _, err := Evaluate(p, nil); err == nil {
			t.Fatalf("year %d accepted", year)
		}
	}
}

func TestRunningTotalOverflowIsRejected(t *testing.T) {
	// Real catalog limits cannot reach int64 overflow, so use a synthetic package.
	p := Package{AssessmentYear: 2025, Version: "synthetic", Status: Draft, Rules: []Rule{
		{ID: "R1", LimitSen: math.MaxInt64}, {ID: "R2", LimitSen: math.MaxInt64},
	}}
	claims := []Claim{
		{ID: "a", RuleID: "R1", AmountSen: math.MaxInt64, ExpenseKey: "k1", EvidenceReference: "x"},
		{ID: "b", RuleID: "R2", AmountSen: math.MaxInt64, ExpenseKey: "k2", EvidenceReference: "x"},
	}
	if _, err := Evaluate(p, claims); err == nil {
		t.Fatal("overflowing total accepted")
	}
}

func TestDraftCatalogNeverProducesApprovedAmounts(t *testing.T) {
	p := YA2025Draft()
	for i, rule := range p.Rules {
		c := claimFor(t, p, "a", rule.ID, 1)
		c.ExpenseKey = fmt.Sprintf("k%d", i)
		r, err := Evaluate(p, []Claim{c})
		if err != nil {
			t.Fatalf("%s: %v", rule.ID, err)
		}
		if r.ApprovedTotalSen != 0 || r.Decisions[0].ApprovedAmountSen != 0 || r.FilingReady || r.Decisions[0].Status != "REVIEW_REQUIRED" {
			t.Fatalf("%s leaked approval: %+v", rule.ID, r)
		}
	}
}
```

- [ ] **Step 5: Run everything**

```bash
gofmt -l internal/policy internal/assessment
go vet ./... && go test ./internal/policy ./internal/assessment ./internal/eligibility -count=1
```

Expected: `gofmt -l` prints only files that were already unformatted before this task (do not reformat untouched files; `registry.go`, `approval.go` and old test files may appear — leave them); `go vet` clean; all three packages `ok`, including `TestUnifiedEngineMatchesRecordedEligibilityBehaviour` with every subtest passing. If a recorded case fails, fix the engine, never the JSON.

- [ ] **Step 6: Commit**

```bash
gofmt -w internal/policy/engine.go internal/policy/equivalence_test.go internal/policy/engine_unified_test.go internal/policy/engine_test.go
git add internal/policy internal/assessment
git commit -m "feat(policy): one Evaluate with exclusive families, expense keys and overflow guards"
```

---

### Task 5: Serve eligibility preview from policy and delete the old package

**Files:**
- Modify: `services/api/cmd/server/main.go`
- Modify: `services/api/internal/policy/catalog.go` (comment)
- Create: `services/api/internal/policy/firsthome_cap_test.go`
- Delete: `services/api/internal/eligibility/`

- [ ] **Step 1: Port the first-home tests that only the old package had**

Create `services/api/internal/policy/firsthome_cap_test.go`:

```go
package policy

import "testing"

func TestFirstHomeEachRequiredFactBlocks(t *testing.T) {
	p := YA2025Draft()
	for _, id := range []string{"FIRST_HOME_LOW", "FIRST_HOME_MID"} {
		for _, fact := range findRule(t, p, id).RequiredFacts {
			c := claimFor(t, p, "a", id, 300000)
			c.Facts[fact] = false
			r, err := Evaluate(p, []Claim{c})
			if err != nil || r.ProvisionalTotalSen != 0 || r.Decisions[0].Status != "REVIEW_REQUIRED" {
				t.Fatalf("%s/%s must block: %+v %v", id, fact, r, err)
			}
		}
	}
}

func TestFirstHomeCapBoundaries(t *testing.T) {
	p := YA2025Draft()
	cases := []struct {
		rule         string
		amount, want int64
	}{
		{"FIRST_HOME_LOW", 699999, 699999}, {"FIRST_HOME_LOW", 700000, 700000}, {"FIRST_HOME_LOW", 700001, 700000},
		{"FIRST_HOME_MID", 499999, 499999}, {"FIRST_HOME_MID", 500000, 500000}, {"FIRST_HOME_MID", 500001, 500000},
	}
	for _, tc := range cases {
		r, err := Evaluate(p, []Claim{claimFor(t, p, "a", tc.rule, tc.amount)})
		if err != nil || r.ProvisionalTotalSen != tc.want || r.ApprovedTotalSen != 0 || r.FilingReady {
			t.Fatalf("%s %d: got %d err=%v", tc.rule, tc.amount, r.ProvisionalTotalSen, err)
		}
	}
}
```

- [ ] **Step 2: Switch the endpoint**

In `services/api/cmd/server/main.go`: delete the import line `"github.com/fikrimastor/mytax/services/api/internal/eligibility"`, and change the handler body to:

```go
 e.POST("/v1/eligibility/preview",func(c echo.Context)error{
  var input struct {AssessmentYear int `json:"assessment_year"`; Claims []policy.Claim `json:"claims"`}
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  if input.AssessmentYear!=2025{return c.JSON(http.StatusBadRequest,map[string]string{"error":"only YA 2025 supported"})}
  result,err:=policy.Evaluate(policy.YA2025Draft(),input.Claims)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
```

In `services/api/internal/policy/catalog.go` the comment mentioning `catalog_parity_test.go` was already replaced in Task 2; confirm with `grep -n parity internal/policy/catalog.go` (no output expected).

- [ ] **Step 3: Delete the old package**

```bash
git rm -r internal/eligibility
grep -rn "internal/eligibility" --include='*.go' . ; echo "exit=$?"
```

Expected: `grep` finds nothing (`exit=1`).

- [ ] **Step 4: Build and test**

```bash
go build ./... && go vet ./... && go test ./... -count=1 2>&1 | tail -20
```

Expected: build and vet clean; every package `ok` (persistence tests skip without a database; that is expected here and is verified in Task 7).

- [ ] **Step 5: Smoke-test the live endpoint**

```bash
PORT=18080 go run ./cmd/server > /tmp/mytax-smoke.log 2>&1 &
sleep 4
curl -s -X POST localhost:18080/v1/eligibility/preview -H 'content-type: application/json' \
  -d '{"assessment_year":2025,"claims":[{"id":"a","rule_id":"MEDICAL_DENTAL","amount_sen":50000,"expense_key":"inv-1","evidence_reference":"doc","facts":{}}]}'
kill %1
```

Expected: JSON with `"status":"REVIEW_REQUIRED"`, `"provisional_amount_sen":0`, reasons containing `UNVERIFIED_FACT:resident`, and `"filing_ready":false`. If the server listens on a different port variable name, read `main.go` for the real one and adjust; report what you used.

- [ ] **Step 6: Commit**

```bash
gofmt -l internal/policy/firsthome_cap_test.go
git add -A internal cmd
git commit -m "refactor: serve eligibility preview from policy and delete internal/eligibility"
```

---

### Task 6: Docs and spec wording

**Files:**
- Modify: `AGENTS.md`, `docs/RELIEF_LIFESTYLE_YA2025.md`, `docs/ASSESSMENT_PREVIEW.md`, `docs/MILESTONE_9.md`, `docs/ROADMAP.md`, `docs/superpowers/specs/2026-11-03-unify-relief-engine-design.md`

- [ ] **Step 1: `AGENTS.md`**

Delete the line `- \`services/api/internal/eligibility\`: expense candidate screening, shared caps, exclusions.` and replace the `policy` line with:

```
- `services/api/internal/policy`: draft relief rules, registry, approval contracts, the canonical fact list and the single claim-evaluation engine (shared caps, exclusive families, duplicate detection).
```

- [ ] **Step 2: `docs/RELIEF_LIFESTYLE_YA2025.md`**

Replace line 27 with `` The rules live only in `internal/policy` (`YA2025Draft`); `internal/eligibility` was removed on 2026-11-03. Amounts are integer sen. ``
Replace line 32 with `- Same \`expense_key\` cannot be claimed twice. This also stops one course fee being claimed as lifestyle and as education up-skilling.`
Replace line 34 with `- Only YA 2025 is accepted by the engine.`
In line 37 replace `` `internal/eligibility/lifestyle_test.go` and `internal/policy/lifestyle_test.go` `` with `` `internal/policy/lifestyle_test.go`, `engine_unified_test.go` and `equivalence_test.go` ``.
Delete line 58 (the note that `policy` relies on a caller `Duplicate` flag); it is no longer true.

- [ ] **Step 3: `docs/ASSESSMENT_PREVIEW.md`**

Append:

````markdown

## Claim shape (changed 2026-11-03)

`claims` use the same shape as `POST /v1/eligibility/preview`:

```json
{"id":"c1","rule_id":"MEDICAL_VACCINATION","amount_sen":50000,"expense_key":"inv-123","evidence_reference":"doc-9","facts":{"resident":true}}
```

`evidence_present` and `duplicate` no longer exist: the server derives duplicates from `expense_key` and checks that `evidence_reference` is set. Decisions use `claim_id`, `status`, `provisional_amount_sen`, `approved_amount_sen` and `reasons`; the result adds `provisional_total_sen` and `filing_ready` (always false).
````

- [ ] **Step 4: `docs/MILESTONE_9.md`**

Add as a new line 3: `**Update 2026-11-03:** this endpoint is now served by \`internal/policy\`. Response fields follow the policy shape (\`claim_id\`, \`status\`, \`provisional_amount_sen\`, \`approved_amount_sen\`, \`provisional_total_sen\`, \`filing_ready\`); rule ids \`EDUCATION_UPSKILL\` and \`EDUCATION_OTHER\` became \`EDUCATION_SELF_SKILLS\` and \`EDUCATION_SELF_OTHER\`.`

- [ ] **Step 5: `docs/ROADMAP.md`**

Replace the two open items `- [ ] Fact vocabulary is inconsistent across rules …` and `- [ ] Two engines (\`policy\` and \`eligibility\`) carry overlapping rules. Decide one source of truth.` with:

```
- [x] One engine and one rule catalog: `internal/policy` (2026-11-03). `internal/eligibility` was removed; see `docs/superpowers/specs/2026-11-03-unify-relief-engine-design.md`.
- [x] Canonical fact list in `internal/policy/facts.go`; a test fails if any rule requires a fact outside it. New rules must use it.
```

In "Suggested order", change step 2 to `2. Rule coverage for the most common receipt categories on the unified engine (medical batch first).`

- [ ] **Step 6: Spec wording**

In the spec, change the sentence that says canonical names are "constants" to say they are "a single sorted list (`CanonicalFacts`)", and in the testing section change the golden case "total overflow" to "claim count limit", adding after item 2: `The int64 running-total overflow guard cannot be reached with real catalog limits, so it is covered by a test on a synthetic package.`

- [ ] **Step 7: Check and commit**

```bash
grep -rn "internal/eligibility" AGENTS.md docs README.md | grep -v "superpowers/"
git add AGENTS.md docs
git commit -m "docs: describe the unified relief engine and new claim shape"
```

Expected: the `grep` prints nothing.

---

### Task 7: Final verification, push and pull request (controller, not a subagent)

- [ ] **Step 1: Full local test with PostgreSQL**

Spawn an `e2e-runner` with `Bash(go:*)`, `Bash(docker:*)`, `Read`, `Grep`, `Glob` allowed. Ask it to: run `go vet ./...` and `go test -race -count=1 ./...`; start a throwaway `postgres:16`, apply `testdata/security_fixture.sql` as `.github/workflows/api.yml` does, run `./internal/persistence/...` with the CI variables and report zero `--- SKIP`; run `go mod tidy` and confirm `git diff --exit-code go.mod go.sum`; remove its container. Read its log before claiming anything.

- [ ] **Step 2: Push and watch CI**

```bash
git push git@github.com:fikrimastor/mytax.git refactor/unify-relief-engine:refactor/unify-relief-engine
gh run list --branch refactor/unify-relief-engine --limit 1
```

Wait until the run completes and is green (the Keycloak live tests run in CI).

- [ ] **Step 3: Ask the user before opening the PR**

Report: commits, verified results by level (committed, locally tested, CI), and the API changes. Open the pull request only when the user asks.

---

## Self-review (done before handing off)

- **Spec coverage:** structure and deletion (Tasks 4, 5); claim and response shape and reason codes (Task 4); merged rules and facts (Tasks 2, 3); API changes (Tasks 4, 5, 6); golden tests and allowed differences (Tasks 1, 4); no new Approved path (`TestDraftCatalogNeverProducesApprovedAmounts`); overflow guard (synthetic package test); docs and `AGENTS.md` (Task 6); verification (Task 7).
- **Placeholders:** none; every code step has code and every command has an expected result.
- **Type consistency:** `Claim` fields `ExpenseKey` and `EvidenceReference`, `Result.ProvisionalTotalSen`, `Rule.ExclusiveFamily`, `CanonicalFacts`/`IsCanonicalFact`, and helpers `claimFor`/`findRule` (defined in `lifestyle_test.go`) are used consistently across tasks.
- **Known judgement calls for the executor to report, not decide:** if a recorded case needs a behaviour change to pass, stop and report instead of editing `relief_cases.json`; if `go vet` reveals another caller of the old claim fields, list it and fix only the field names.
