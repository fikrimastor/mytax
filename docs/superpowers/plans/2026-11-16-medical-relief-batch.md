# Medical relief batch (YA 2025) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Draft YA 2025 rules for Act 53 s.46(1)(c), (d), (e), (g)(i), (g)(ii), (h), (ha) to `internal/policy`, with fixed-amount rules and nested groups, backed by golden fixtures from LHDN Public Ruling No. 7/2025.

**Architecture:** One shared structure check (`validateStructure`) feeds both `Evaluate` and `ValidateRelease`. `Rule.FixedAmount` makes the engine use `LimitSen` as the claimed amount. `Group.Parent` makes a claim's allowance the minimum over its group and every ancestor. Rules and their canonical facts land together in `catalog.go` and `facts.go`.

**Tech Stack:** Go 1.25 (`services/api`), standard `testing`, JSON fixtures under `internal/policy/testdata`.

**Spec:** `docs/superpowers/specs/2026-11-07-medical-relief-batch-design.md` (approved 2026-11-16).

## Global Constraints

- Every rule stays **Draft**. Nothing in this plan creates an Approved package. `Status: Approved` in tests is an arithmetic fixture only and must say so in a comment.
- Amounts are integer sen. RM1 = 100 sen.
- All rule limits are per taxpayer. One `Evaluate` call per taxpayer.
- Allocation order stays claim-ID order.
- YA 2025 values only. Do not use any Finance Act 2025 (YA 2026) value: learning disability stays RM6,000, not RM10,000.
- Public repo: no real taxpayer data. Fixture names come from PR 7/2025 worked examples only.
- Work from `services/api`. Before each commit run `gofmt -w` on the Go files that task touched (code blocks here are not guaranteed gofmt-aligned). Do not run the full suite mid-task beyond the commands each task lists; the final task runs everything once.
- Commit only the files the task lists.

## File map

| File | Task | Change |
|---|---|---|
| `services/api/internal/policy/engine.go` | 1 | `Rule.FixedAmount`, `Group.Parent`; `Evaluate` uses `validateStructure`, fixed amount and the group chain |
| `services/api/internal/policy/structure.go` (new) | 1 | `validateStructure`, `groupChain` |
| `services/api/internal/policy/approval.go` | 1 | `ValidateRelease` calls `validateStructure` |
| `services/api/internal/policy/engine_nested_test.go` (new) | 1 | nested-group allocation and validation |
| `services/api/internal/policy/engine_fixed_test.go` (new) | 1 | fixed-amount behaviour and validation |
| `services/api/internal/policy/facts.go` | 2 | 18 new canonical facts |
| `services/api/internal/policy/catalog.go` | 2 | `PR72025URL`, `medicalSources`, 3 groups, 12 rules |
| `services/api/internal/policy/engine_unified_test.go` | 2 | fixed rules claim 0 in `TestDraftCatalogNeverProducesApprovedAmounts` |
| `services/api/internal/policy/testdata/pr7_2025_cases.json` (new) | 2 | golden fixtures |
| `services/api/internal/policy/medical_test.go` (new) | 2 | fixture replay, boundaries, fail-closed, real-catalog nesting |
| `services/api/internal/policy/registry.go` | 3 | notes cite PR 7/2025 |
| `docs/RELIEF_MEDICAL_YA2025.md` (new) | 3 | rule and source doc |
| `docs/ROADMAP.md`, `docs/ASSESSMENT_PREVIEW.md` | 3 | status and fixed-amount claim shape |

Task order: Task 1 → Task 2. Task 3 touches no Go file in Tasks 1–2 and can run beside them. Task 4 runs last.

---

### Task 1: Engine — fixed-amount rules and nested groups

**Files:**
- Create: `services/api/internal/policy/structure.go`
- Modify: `services/api/internal/policy/engine.go:29-41` (types), `:106-130` (setup), `:146-216` (claim loop)
- Modify: `services/api/internal/policy/approval.go:17-28`
- Test: `services/api/internal/policy/engine_nested_test.go`, `services/api/internal/policy/engine_fixed_test.go`

**Interfaces:**
- Consumes: existing `Package`, `Rule`, `Group`, `Claim`, `Evaluate`, `ValidateRelease`, `ReleaseRecord`.
- Produces:
  - `Rule.FixedAmount bool` with tag `json:"fixed_amount,omitempty"`
  - `Group.Parent string` with tag `json:"parent,omitempty"`
  - `func validateStructure(p Package) (map[string]Rule, map[string]Group, error)`
  - `func groupChain(groups map[string]Group, id string) []string`
  - Contract: a claim on a `FixedAmount` rule must send `AmountSen == 0`; any other value makes `Evaluate` return an error. A fixed rule with `LimitSen == 0` or a non-empty `Group` is rejected by both `Evaluate` and `ValidateRelease`.

- [ ] **Step 1: Write the failing nested-group tests**

Create `services/api/internal/policy/engine_nested_test.go`:

```go
package policy

import "testing"

// Status Approved in this file is an arithmetic fixture, not tax-policy sign-off.

var testSources = []Source{{URL: "https://www.hasil.gov.my/test", Section: "synthetic", RetrievedOn: "2026-11-16"}}

func testRecord(p Package) ReleaseRecord {
	return ReleaseRecord{PackageVersion: p.Version, AssessmentYear: p.AssessmentYear, ReviewerID: "r", ReviewedAt: "2026-11-16",
		ApprovalReference: "synthetic", SourceSnapshotSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
}

func nestedPackage() Package {
	return Package{AssessmentYear: 2025, Version: "t", Status: Approved,
		Groups: []Group{{ID: "TOP", LimitSen: 1000}, {ID: "SUB", Parent: "TOP", LimitSen: 300}},
		Rules: []Rule{
			{ID: "BIG", Group: "TOP", LimitSen: 1000, Sources: testSources},
			{ID: "S1", Group: "SUB", LimitSen: 300, Sources: testSources},
			{ID: "S2", Group: "SUB", LimitSen: 300, Sources: testSources},
		}}
}

func synthClaim(id, rule string, amt int64) Claim {
	return Claim{ID: id, RuleID: rule, AmountSen: amt, ExpenseKey: "exp-" + id, EvidenceReference: "x"}
}

func wantAmounts(t *testing.T, r Result, want []int64, total int64) {
	t.Helper()
	if len(r.Decisions) != len(want) {
		t.Fatalf("%d decisions, want %d", len(r.Decisions), len(want))
	}
	for i, w := range want {
		if got := r.Decisions[i].ProvisionalAmountSen; got != w {
			t.Fatalf("claim %s = %d, want %d (%v)", r.Decisions[i].ClaimID, got, w, r.Decisions[i].Reasons)
		}
	}
	if r.ProvisionalTotalSen != total {
		t.Fatalf("total %d, want %d", r.ProvisionalTotalSen, total)
	}
}

func hasReason(d Decision, reason string) bool {
	for _, r := range d.Reasons {
		if r == reason {
			return true
		}
	}
	return false
}

func TestChildGroupCapsItsRulesAndParentKeepsTheRest(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "S1", 200), synthClaim("b", "S2", 200), synthClaim("c", "BIG", 900)})
	if err != nil {
		t.Fatal(err)
	}
	// SUB is full after 300; TOP has 700 left for BIG.
	wantAmounts(t, r, []int64{200, 100, 700}, 1000)
	if !hasReason(r.Decisions[1], "LIMIT_REACHED") || !hasReason(r.Decisions[2], "LIMIT_REACHED") {
		t.Fatalf("missing LIMIT_REACHED: %+v", r.Decisions)
	}
}

func TestFullChildGroupGivesZeroWhileParentStillPays(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "S1", 300), synthClaim("b", "S2", 50), synthClaim("c", "BIG", 1000)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{300, 0, 700}, 1000)
}

func TestParentGroupCapBindsChildRules(t *testing.T) {
	r, err := Evaluate(nestedPackage(), []Claim{synthClaim("a", "BIG", 900), synthClaim("b", "S1", 300)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{900, 100}, 1000)
}

func TestNestedGroupStructureIsValidated(t *testing.T) {
	if err := ValidateRelease(nestedPackage(), testRecord(nestedPackage())); err != nil {
		t.Fatalf("valid nested package rejected: %v", err)
	}
	cases := map[string]func(*Package){
		"unknown parent":            func(p *Package) { p.Groups[1].Parent = "NOPE" },
		"self parent":               func(p *Package) { p.Groups[0].Parent = "TOP" },
		"two-group cycle":           func(p *Package) { p.Groups[0].Parent = "SUB"; p.Groups[1].LimitSen = 1000 },
		"child limit above parent":  func(p *Package) { p.Groups[1].LimitSen = 1001 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := nestedPackage()
			mutate(&p)
			if _, err := Evaluate(p, nil); err == nil {
				t.Fatal("Evaluate accepted the package")
			}
			if err := ValidateRelease(p, testRecord(p)); err == nil {
				t.Fatal("ValidateRelease accepted the package")
			}
		})
	}
}
```

- [ ] **Step 2: Write the failing fixed-amount tests**

Create `services/api/internal/policy/engine_fixed_test.go`:

```go
package policy

import "testing"

// Status Approved in this file is an arithmetic fixture, not tax-policy sign-off.

func fixedPackage() Package {
	return Package{AssessmentYear: 2025, Version: "t", Status: Approved,
		Groups: []Group{{ID: "G", LimitSen: 1000}},
		Rules: []Rule{
			{ID: "FIX", FixedAmount: true, LimitSen: 700000, Sources: testSources},
			{ID: "EXP", Group: "G", LimitSen: 1000, Sources: testSources},
		}}
}

func TestFixedAmountClaimReceivesTheRuleLimit(t *testing.T) {
	r, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", 0)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000}, 700000)
	if hasReason(r.Decisions[0], "LIMIT_REACHED") {
		t.Fatalf("fixed claim must not be LIMIT_REACHED: %+v", r.Decisions[0])
	}
}

func TestFixedAmountClaimRejectsAnySentAmount(t *testing.T) {
	for _, amt := range []int64{1, -1, 700000} {
		if _, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", amt)}); err == nil {
			t.Fatalf("amount %d accepted on a fixed-amount rule", amt)
		}
	}
}

func TestExpenseRuleStillRejectsZeroAmount(t *testing.T) {
	if _, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "EXP", 0)}); err == nil {
		t.Fatal("zero amount accepted on an expense rule")
	}
}

func TestSecondFixedAmountClaimGetsZero(t *testing.T) {
	r, err := Evaluate(fixedPackage(), []Claim{synthClaim("a", "FIX", 0), synthClaim("b", "FIX", 0)})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000, 0}, 700000)
	if !hasReason(r.Decisions[1], "LIMIT_REACHED") {
		t.Fatalf("second fixed claim: %+v", r.Decisions[1])
	}
}

func TestFixedAmountRuleStructureIsValidated(t *testing.T) {
	if err := ValidateRelease(fixedPackage(), testRecord(fixedPackage())); err != nil {
		t.Fatalf("valid fixed package rejected: %v", err)
	}
	cases := map[string]func(*Package){
		"zero limit": func(p *Package) { p.Rules[0].LimitSen = 0 },
		"in a group": func(p *Package) { p.Rules[0].Group = "G" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := fixedPackage()
			mutate(&p)
			if _, err := Evaluate(p, nil); err == nil {
				t.Fatal("Evaluate accepted the package")
			}
			if err := ValidateRelease(p, testRecord(p)); err == nil {
				t.Fatal("ValidateRelease accepted the package")
			}
		})
	}
}
```

- [ ] **Step 3: Run the new tests and confirm they fail to compile**

Run: `cd services/api && go test ./internal/policy/ -run 'Nested|Fixed|ChildGroup|ParentGroup|ExpenseRuleStill|SecondFixed' -count=1`
Expected: FAIL, `unknown field Parent in struct literal of type Group` and `unknown field FixedAmount`.

- [ ] **Step 4: Add the type fields**

In `engine.go`, replace the `Rule` and `Group` types (lines 29–41) with:

```go
type Rule struct {
	ID              string `json:"id"`
	Group           string `json:"group"`
	ExclusiveFamily string `json:"exclusive_family,omitempty"`
	// FixedAmount rules grant LimitSen in full; claims must send amount_sen 0.
	FixedAmount   bool     `json:"fixed_amount,omitempty"`
	LimitSen      int64    `json:"limit_sen"`
	RequiredFacts []string `json:"required_facts"`
	Sources       []Source `json:"sources"`
}

type Group struct {
	ID string `json:"id"`
	// Parent names an enclosing group whose limit also applies to this group's claims.
	Parent   string `json:"parent,omitempty"`
	LimitSen int64  `json:"limit_sen"`
}
```

- [ ] **Step 5: Create `structure.go`**

```go
package policy

import "errors"

// validateStructure checks the rule and group graph. Evaluate and ValidateRelease both use it,
// so a package that cannot be evaluated can never be released.
func validateStructure(p Package) (map[string]Rule, map[string]Group, error) {
	groups := map[string]Group{}
	for _, g := range p.Groups {
		if g.ID == "" || g.LimitSen < 0 {
			return nil, nil, errors.New("invalid group")
		}
		if _, ok := groups[g.ID]; ok {
			return nil, nil, errors.New("duplicate group")
		}
		groups[g.ID] = g
	}
	for _, g := range p.Groups {
		if g.Parent == "" {
			continue
		}
		parent, ok := groups[g.Parent]
		if !ok {
			return nil, nil, errors.New("unknown parent group")
		}
		if g.LimitSen > parent.LimitSen {
			return nil, nil, errors.New("child group limit exceeds parent limit")
		}
		seen := map[string]bool{g.ID: true}
		for id := g.Parent; id != ""; id = groups[id].Parent {
			if seen[id] {
				return nil, nil, errors.New("group parent cycle")
			}
			seen[id] = true
		}
	}
	rules := map[string]Rule{}
	for _, rule := range p.Rules {
		if rule.ID == "" || rule.LimitSen < 0 {
			return nil, nil, errors.New("invalid rule")
		}
		if _, ok := rules[rule.ID]; ok {
			return nil, nil, errors.New("duplicate rule")
		}
		if rule.Group != "" {
			if _, ok := groups[rule.Group]; !ok {
				return nil, nil, errors.New("unknown group")
			}
		}
		if rule.FixedAmount && (rule.LimitSen == 0 || rule.Group != "") {
			return nil, nil, errors.New("fixed-amount rule needs a positive limit and no group")
		}
		rules[rule.ID] = rule
	}
	return rules, groups, nil
}

// groupChain returns id followed by every ancestor group. validateStructure has ruled out cycles.
func groupChain(groups map[string]Group, id string) []string {
	var chain []string
	for ; id != ""; id = groups[id].Parent {
		chain = append(chain, id)
	}
	return chain
}
```

- [ ] **Step 6: Use it in `Evaluate`**

In `engine.go` replace lines 106–130 (the two maps and both setup loops) with:

```go
	rules, groups, err := validateStructure(p)
	if err != nil {
		return Result{}, err
	}
```

Replace the claim-loop head (lines 147–149):

```go
		if c.ID == "" || c.AmountSen <= 0 {
			return Result{}, errors.New("claim ID and positive amount required")
		}
```

with:

```go
		rule, ok := rules[c.RuleID]
		claimed := c.AmountSen
		if ok && rule.FixedAmount {
			if c.AmountSen != 0 {
				return Result{}, errors.New("fixed-amount relief takes amount_sen 0")
			}
			claimed = rule.LimitSen
		}
		if c.ID == "" || claimed <= 0 {
			return Result{}, errors.New("claim ID and positive amount required")
		}
```

Delete the later `rule, ok := rules[c.RuleID]` line (old line 155); keep the `if !ok { ... UNKNOWN_RELIEF_RULE }` block after it.

Replace the allocation block (old lines 181–202, from `available := ...` through the `usedGroups` update) with:

```go
			available := rule.LimitSen - usedRules[rule.ID]
			chain := groupChain(groups, rule.Group)
			for _, gid := range chain {
				if g := groups[gid].LimitSen - usedGroups[gid]; g < available {
					available = g
				}
			}
			if available < 0 {
				available = 0
			}
			allowed := claimed
			if allowed > available {
				allowed = available
			}
			if res.ProvisionalTotalSen > math.MaxInt64-allowed {
				return Result{}, errors.New("overflow")
			}
			d.ProvisionalAmountSen = allowed
			res.ProvisionalTotalSen += allowed
			usedRules[rule.ID] += allowed
			for _, gid := range chain {
				usedGroups[gid] += allowed
			}
```

Change `if allowed < c.AmountSen {` (old line 213) to `if allowed < claimed {`.

Add to the `Evaluate` doc comment, after the "Allocation order" paragraph:

```go
// Nested groups: a claim is limited by its rule, its group and every ancestor group, and its
// allowed amount is charged to all of them. When a group cap binds, the per-line split
// depends on claim-ID order; the group total does not.
//
// Fixed amounts: a claim on a FixedAmount rule sends amount_sen 0 and is granted LimitSen.
```

- [ ] **Step 7: Use it in `ValidateRelease`**

In `approval.go`, after `if len(p.Rules)==0{return errors.New("empty release")}` add:

```go
 if _,_,err:=validateStructure(p);err!=nil{return err}
```

- [ ] **Step 8: Run the policy package tests**

Run: `cd services/api && go test ./internal/policy/ -count=1`
Expected: PASS, including `TestSetupValidationErrors`, `TestUnifiedEngineMatchesRecordedEligibilityBehaviour` and `TestRuleLimitAndGroupLimitBothApply` unchanged.

- [ ] **Step 9: Commit**

```bash
git add services/api/internal/policy/engine.go services/api/internal/policy/structure.go services/api/internal/policy/approval.go services/api/internal/policy/engine_nested_test.go services/api/internal/policy/engine_fixed_test.go
git commit -m "feat(policy): fixed-amount rules and nested relief groups"
```

---

### Task 2: YA 2025 medical rules, facts and PR 7/2025 golden fixtures

**Files:**
- Modify: `services/api/internal/policy/facts.go:9-35`
- Modify: `services/api/internal/policy/catalog.go` (constants, `YA2025Draft`)
- Modify: `services/api/internal/policy/engine_unified_test.go:79-94`
- Create: `services/api/internal/policy/testdata/pr7_2025_cases.json`
- Create: `services/api/internal/policy/medical_test.go`

**Interfaces:**
- Consumes (Task 1): `Rule.FixedAmount`, `Group.Parent`; `Evaluate` grants `LimitSen` for `amount_sen 0` on fixed rules. Test helpers from Task 1: `wantAmounts(t, r Result, want []int64, total int64)`, `hasReason(d Decision, reason string) bool`. Existing helpers: `claimFor(t, p, id, rule string, amt int64) Claim` (fills every required fact true, `ExpenseKey "exp-"+id`), `findRule`, `recordedClaim`, `recordedCase`, `buildRecordedClaims(p Package, c recordedCase) []Claim` (fills required facts, then deletes `missing_facts` and sets `false_facts` to false).
- Produces: `const PR72025URL string`; `func medicalSources(act, ruling string) []Source`; rule IDs `PARENTS_MEDICAL`, `PARENTS_CARER`, `PARENTS_CHECKUP`, `DISABILITY_EQUIPMENT`, `DISABLED_SELF`, `MEDICAL_SERIOUS_DISEASE`, `MEDICAL_FERTILITY`, `MEDICAL_SCREENING_EXAM`, `MEDICAL_SCREENING_DETECTION_TEST`, `MEDICAL_SCREENING_SELF_TEST_DEVICE`, `MEDICAL_SCREENING_MENTAL_HEALTH`, `MEDICAL_LEARNING_DISABILITY`; group IDs `PARENTS`, `DISABILITY_EQUIPMENT`, `MEDICAL_SCREENING`.

- [ ] **Step 1: Write the golden fixtures**

Create `services/api/internal/policy/testdata/pr7_2025_cases.json`. Each case is one taxpayer and one `Evaluate` call. `encoding` records choices the ruling does not state.

```json
[
  {"example": 3, "taxpayer": "Zubir", "encoding": "Grandparents' medical and parents' dental are separate PARENTS_MEDICAL expenses; vaccination is PARENTS_CHECKUP (complete examination includes vaccination, FA 2024 s.9(a)).",
   "claims": [
     {"id": "a", "rule_id": "PARENTS_MEDICAL", "amount_sen": 450000, "expense_key": "ex3-grandparents-medical", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "PARENTS_MEDICAL", "amount_sen": 100000, "expense_key": "ex3-parents-dental", "evidence_reference": "receipt"},
     {"id": "c", "rule_id": "PARENTS_CHECKUP", "amount_sen": 280000, "expense_key": "ex3-vaccination", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 450000}, {"id": "b", "amount_sen": 100000}, {"id": "c", "amount_sen": 100000, "reason": "LIMIT_REACHED"}], "total_sen": 650000}},
  {"example": 4, "taxpayer": "Megat", "encoding": "The ruling gives no price; 300000 sen is chosen.",
   "claims": [{"id": "a", "rule_id": "DISABILITY_EQUIPMENT", "amount_sen": 300000, "expense_key": "ex4-wheelchair", "evidence_reference": "receipt", "false_facts": ["disability_certified_by_jkm"]}],
   "expect": {"lines": [{"id": "a", "amount_sen": 0, "reason": "UNVERIFIED_FACT:disability_certified_by_jkm"}], "total_sen": 0}},
  {"example": 8, "taxpayer": "Jasmine",
   "claims": [{"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 1000000, "expense_key": "ex8-jasmine", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 1000000}], "total_sen": 1000000}},
  {"example": 8, "taxpayer": "Isaac",
   "claims": [{"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 400000, "expense_key": "ex8-isaac", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 400000}], "total_sen": 400000}},
  {"example": 9, "taxpayer": "Fahim",
   "claims": [{"id": "a", "rule_id": "MEDICAL_FERTILITY", "amount_sen": 1000000, "expense_key": "ex9-ivf-fahim", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 1000000}], "total_sen": 1000000}},
  {"example": 9, "taxpayer": "Anita", "encoding": "The receipt is in Fahim's name, so paid_by_claimant is false for Anita.",
   "claims": [{"id": "a", "rule_id": "MEDICAL_FERTILITY", "amount_sen": 1000000, "expense_key": "ex9-ivf-anita", "evidence_reference": "receipt", "false_facts": ["paid_by_claimant"]}],
   "expect": {"lines": [{"id": "a", "amount_sen": 0, "reason": "UNVERIFIED_FACT:paid_by_claimant"}], "total_sen": 0}},
  {"example": 10, "taxpayer": "Fahim",
   "claims": [{"id": "a", "rule_id": "MEDICAL_FERTILITY", "amount_sen": 1000000, "expense_key": "ex10-fahim", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 1000000}], "total_sen": 1000000}},
  {"example": 10, "taxpayer": "Anita",
   "claims": [{"id": "a", "rule_id": "MEDICAL_FERTILITY", "amount_sen": 1000000, "expense_key": "ex10-anita", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 1000000}], "total_sen": 1000000}},
  {"example": 11, "taxpayer": "Azli", "encoding": "Claim IDs follow the ruling's table order (serious disease, then fertility); the ruling states only the RM10,000 total.",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 600000, "expense_key": "ex11-son-serious", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_FERTILITY", "amount_sen": 1200000, "expense_key": "ex11-fertility", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 600000}, {"id": "b", "amount_sen": 400000, "reason": "LIMIT_REACHED"}], "total_sen": 1000000}},
  {"example": 13, "taxpayer": "Khadijah",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 50000, "expense_key": "ex13-influenza", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_DENTAL", "amount_sen": 140000, "expense_key": "ex13-dental", "evidence_reference": "receipt"},
     {"id": "c", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 500000, "expense_key": "ex13-serious", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 50000}, {"id": "b", "amount_sen": 100000, "reason": "LIMIT_REACHED"}, {"id": "c", "amount_sen": 500000}], "total_sen": 650000}},
  {"example": 14, "taxpayer": "John", "encoding": "Ruling states only the RM1,000 total; claim IDs follow the text order.",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_SCREENING_MENTAL_HEALTH", "amount_sen": 80000, "expense_key": "ex14-psychology", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_SCREENING_EXAM", "amount_sen": 80000, "expense_key": "ex14-exam", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 80000}, {"id": "b", "amount_sen": 20000, "reason": "LIMIT_REACHED"}], "total_sen": 100000}},
  {"example": 16, "taxpayer": "Diana", "encoding": "The ruling lists 12,500 (serious disease 12,000 plus additional vaccination 500) as one line restricted to 10,000 and gives no amount for the screening lines; encoded as four claims in that order.",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 1200000, "expense_key": "ex16-kidney", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 50000, "expense_key": "ex16-vaccination", "evidence_reference": "receipt"},
     {"id": "c", "rule_id": "MEDICAL_SCREENING_EXAM", "amount_sen": 75000, "expense_key": "ex16-checkup", "evidence_reference": "receipt"},
     {"id": "d", "rule_id": "MEDICAL_SCREENING_DETECTION_TEST", "amount_sen": 30000, "expense_key": "ex16-pcr", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 1000000, "reason": "LIMIT_REACHED"}, {"id": "b", "amount_sen": 0, "reason": "LIMIT_REACHED"}, {"id": "c", "amount_sen": 0, "reason": "LIMIT_REACHED"}, {"id": "d", "amount_sen": 0, "reason": "LIMIT_REACHED"}], "total_sen": 1000000}},
  {"example": 17, "taxpayer": "Jasmin", "encoding": "Ruling states the RM10,000 total; per-line split follows claim-ID order.",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 900000, "expense_key": "ex17-serious", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_SCREENING_DETECTION_TEST", "amount_sen": 20000, "expense_key": "ex17-mammogram", "evidence_reference": "receipt"},
     {"id": "c", "rule_id": "MEDICAL_SCREENING_SELF_TEST_DEVICE", "amount_sen": 70000, "expense_key": "ex17-oximeter-bp", "evidence_reference": "receipt"},
     {"id": "d", "rule_id": "MEDICAL_SCREENING_SELF_TEST_DEVICE", "amount_sen": 40000, "expense_key": "ex17-children-kits", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 900000}, {"id": "b", "amount_sen": 20000}, {"id": "c", "amount_sen": 70000}, {"id": "d", "amount_sen": 10000, "reason": "LIMIT_REACHED"}], "total_sen": 1000000}},
  {"example": 18, "taxpayer": "Anis",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 30000, "expense_key": "ex18-influenza", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_LEARNING_DISABILITY", "amount_sen": 700000, "expense_key": "ex18-intervention", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 30000}, {"id": "b", "amount_sen": 600000, "reason": "LIMIT_REACHED"}], "total_sen": 630000}},
  {"example": 19, "taxpayer": "Anis", "encoding": "The ruling's line column (9,800, 300, 6,000) is before the RM10,000 group cap; the engine applies the cap per line in claim-ID order.",
   "claims": [
     {"id": "a", "rule_id": "MEDICAL_SERIOUS_DISEASE", "amount_sen": 980000, "expense_key": "ex19-serious", "evidence_reference": "receipt"},
     {"id": "b", "rule_id": "MEDICAL_VACCINATION", "amount_sen": 30000, "expense_key": "ex19-influenza", "evidence_reference": "receipt"},
     {"id": "c", "rule_id": "MEDICAL_LEARNING_DISABILITY", "amount_sen": 700000, "expense_key": "ex19-intervention", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 980000}, {"id": "b", "amount_sen": 20000, "reason": "LIMIT_REACHED"}, {"id": "c", "amount_sen": 0, "reason": "LIMIT_REACHED"}], "total_sen": 1000000}},
  {"example": 20, "taxpayer": "Peter",
   "claims": [{"id": "a", "rule_id": "MEDICAL_LEARNING_DISABILITY", "amount_sen": 700000, "expense_key": "ex20-peter", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 600000, "reason": "LIMIT_REACHED"}], "total_sen": 600000}},
  {"example": 20, "taxpayer": "Jill",
   "claims": [{"id": "a", "rule_id": "MEDICAL_LEARNING_DISABILITY", "amount_sen": 300000, "expense_key": "ex20-jill", "evidence_reference": "receipt"}],
   "expect": {"lines": [{"id": "a", "amount_sen": 300000}], "total_sen": 300000}}
]
```

- [ ] **Step 2: Write the failing tests**

Create `services/api/internal/policy/medical_test.go`:

```go
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Golden fixtures from LHDN Public Ruling No. 7/2025 Examples 3-20 (excerpt in docs/sources/pr7-2025).
// Examples 12 and 15 need joint assessment (s.46(3)) and are not covered.

type pr7Line struct {
	ID        string `json:"id"`
	AmountSen int64  `json:"amount_sen"`
	Reason    string `json:"reason"`
}

type pr7Case struct {
	Example  int             `json:"example"`
	Taxpayer string          `json:"taxpayer"`
	Claims   []recordedClaim `json:"claims"`
	Expect   struct {
		Lines    []pr7Line `json:"lines"`
		TotalSen int64     `json:"total_sen"`
	} `json:"expect"`
}

func TestPR72025WorkedExamples(t *testing.T) {
	raw, err := os.ReadFile("testdata/pr7_2025_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []pr7Case
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 17 {
		t.Fatalf("%d cases, want 17", len(cases))
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("example %d %s", c.Example, c.Taxpayer), func(t *testing.T) {
			p := YA2025Draft()
			r, err := Evaluate(p, buildRecordedClaims(p, recordedCase{Claims: c.Claims}))
			if err != nil {
				t.Fatal(err)
			}
			want := make([]int64, len(c.Expect.Lines))
			for i, l := range c.Expect.Lines {
				want[i] = l.AmountSen
			}
			wantAmounts(t, r, want, c.Expect.TotalSen)
			for i, l := range c.Expect.Lines {
				d := r.Decisions[i]
				if d.ClaimID != l.ID || d.Status != "REVIEW_REQUIRED" || d.ApprovedAmountSen != 0 {
					t.Fatalf("line %d: %+v", i, d)
				}
				if l.Reason != "" && !hasReason(d, l.Reason) {
					t.Fatalf("line %s lacks %s: %v", l.ID, l.Reason, d.Reasons)
				}
			}
		})
	}
}

func TestMedicalCapBoundaries(t *testing.T) {
	caps := map[string]int64{
		"PARENTS_MEDICAL": 800000, "PARENTS_CARER": 800000, "PARENTS_CHECKUP": 100000,
		"DISABILITY_EQUIPMENT": 600000, "MEDICAL_SERIOUS_DISEASE": 1000000, "MEDICAL_FERTILITY": 1000000,
		"MEDICAL_VACCINATION": 100000, "MEDICAL_DENTAL": 100000, "MEDICAL_LEARNING_DISABILITY": 600000,
		"MEDICAL_SCREENING_EXAM": 100000, "MEDICAL_SCREENING_DETECTION_TEST": 100000,
		"MEDICAL_SCREENING_SELF_TEST_DEVICE": 100000, "MEDICAL_SCREENING_MENTAL_HEALTH": 100000,
	}
	p := YA2025Draft()
	for rule, cap := range caps {
		t.Run(rule, func(t *testing.T) {
			at, err := Evaluate(p, []Claim{claimFor(t, p, "a", rule, cap)})
			if err != nil {
				t.Fatal(err)
			}
			wantAmounts(t, at, []int64{cap}, cap)
			if hasReason(at.Decisions[0], "LIMIT_REACHED") {
				t.Fatalf("at cap flagged: %v", at.Decisions[0].Reasons)
			}
			over, err := Evaluate(p, []Claim{claimFor(t, p, "a", rule, cap+1)})
			if err != nil {
				t.Fatal(err)
			}
			wantAmounts(t, over, []int64{cap}, cap)
			if !hasReason(over.Decisions[0], "LIMIT_REACHED") {
				t.Fatalf("one sen over not flagged: %v", over.Decisions[0].Reasons)
			}
		})
	}
}

func TestMedicalGroupBoundaries(t *testing.T) {
	cases := []struct {
		name         string
		first, other string
		firstAmt     int64
		rest         int64 // amount left in the group after the first claim
	}{
		{"PARENTS RM8,000", "PARENTS_MEDICAL", "PARENTS_CARER", 500000, 300000},
		{"MEDICAL_SELF_FAMILY RM10,000", "MEDICAL_SERIOUS_DISEASE", "MEDICAL_FERTILITY", 900000, 100000},
		{"MEDICAL_SCREENING RM1,000", "MEDICAL_SCREENING_EXAM", "MEDICAL_SCREENING_DETECTION_TEST", 60000, 40000},
	}
	p := YA2025Draft()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, extra := range []int64{0, 1} {
				r, err := Evaluate(p, []Claim{claimFor(t, p, "a", tc.first, tc.firstAmt), claimFor(t, p, "b", tc.other, tc.rest+extra)})
				if err != nil {
					t.Fatal(err)
				}
				wantAmounts(t, r, []int64{tc.firstAmt, tc.rest}, tc.firstAmt+tc.rest)
				if hasReason(r.Decisions[1], "LIMIT_REACHED") != (extra == 1) {
					t.Fatalf("extra %d: %v", extra, r.Decisions[1].Reasons)
				}
			}
		})
	}
}

func TestDisabledSelfIsAFixedRM7000(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "DISABLED_SELF", 0)
	b := claimFor(t, p, "b", "DISABLED_SELF", 0)
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{700000, 0}, 700000)
	if _, err := Evaluate(p, []Claim{claimFor(t, p, "a", "DISABLED_SELF", 700000)}); err == nil {
		t.Fatal("non-zero amount accepted on DISABLED_SELF")
	}
}

func TestFullScreeningGroupLeavesSeriousDiseaseTheRest(t *testing.T) {
	p := YA2025Draft()
	r, err := Evaluate(p, []Claim{
		claimFor(t, p, "a", "MEDICAL_SCREENING_EXAM", 100000),
		claimFor(t, p, "b", "MEDICAL_SCREENING_SELF_TEST_DEVICE", 5000),
		claimFor(t, p, "c", "MEDICAL_SERIOUS_DISEASE", 1000000),
	})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{100000, 0, 900000}, 1000000)
}

func TestSameExpenseCannotBeClaimedUnderParentsAndEquipment(t *testing.T) {
	p := YA2025Draft()
	a := claimFor(t, p, "a", "PARENTS_MEDICAL", 200000)
	b := claimFor(t, p, "b", "DISABILITY_EQUIPMENT", 200000)
	b.ExpenseKey = a.ExpenseKey
	r, err := Evaluate(p, []Claim{a, b})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts(t, r, []int64{200000, 0}, 200000)
	if !hasReason(r.Decisions[1], "DUPLICATE_EXPENSE") {
		t.Fatalf("%v", r.Decisions[1].Reasons)
	}
}

func TestMedicalFailsClosedOnUnverifiedFacts(t *testing.T) {
	cases := []struct {
		rule, fact string
		amt        int64
	}{
		{"DISABILITY_EQUIPMENT", "disability_certified_by_jkm", 100000},
		{"DISABLED_SELF", "disability_certified_by_jkm", 0},
		{"MEDICAL_SERIOUS_DISEASE", "qualifying_serious_disease", 100000}, // disease not named in PR §6.6.2
		{"PARENTS_CHECKUP", "qualifying_examination", 50000},               // laboratory test only, PR §6.7.4
		{"MEDICAL_SCREENING_EXAM", "qualifying_examination", 50000},
		{"MEDICAL_LEARNING_DISABILITY", "child_age_18_or_below", 100000},
		{"MEDICAL_FERTILITY", "qualifying_beneficiary", 100000}, // not self or spouse
	}
	p := YA2025Draft()
	for _, tc := range cases {
		t.Run(tc.rule+"/"+tc.fact, func(t *testing.T) {
			c := claimFor(t, p, "a", tc.rule, tc.amt)
			if _, ok := c.Facts[tc.fact]; !ok {
				t.Fatalf("%s does not require %s", tc.rule, tc.fact)
			}
			c.Facts[tc.fact] = false
			r, err := Evaluate(p, []Claim{c})
			if err != nil {
				t.Fatal(err)
			}
			d := r.Decisions[0]
			if d.Status != "REVIEW_REQUIRED" || d.ProvisionalAmountSen != 0 || !hasReason(d, "UNVERIFIED_FACT:"+tc.fact) {
				t.Fatalf("%+v", d)
			}
		})
	}
}
```

- [ ] **Step 3: Run them and confirm they fail**

Run: `cd services/api && go test ./internal/policy/ -run 'PR72025|Medical|DisabledSelf|FullScreening|SameExpense' -count=1`
Expected: FAIL, `rule PARENTS_MEDICAL missing` (from `findRule`) and Example 3 decisions `UNKNOWN_RELIEF_RULE`.

- [ ] **Step 4: Add the canonical facts**

Replace the `CanonicalFacts` literal in `facts.go` with this sorted list:

```go
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
```

- [ ] **Step 5: Add sources, groups and rules to `catalog.go`**

After `firstHomeSources`, add:

```go
// PR72025URL is the official IRBM URL of LHDN Public Ruling No. 7/2025 (7th edition, 5 December 2025).
// It returned HTTP 500 on 2026-11-07 and 2026-11-16; the text used is the MAICSA mirror recorded,
// with its SHA-256, in docs/sources/pr7-2025. A ruling is an administrative interpretation (s.138A), not law.
const PR72025URL = "https://www.hasil.gov.my/wp-content/uploads/pr-7-2025.pdf"

// medicalSources cites the Act 53 provision for YA 2025 and the PR 7/2025 paragraph that interprets it.
func medicalSources(act, ruling string) []Source {
	return []Source{
		{URL: Act53URL, Section: "Income Tax Act 1967 (Act 53) updated text as at 21 May 2024, " + act + "; not an authentic text", RetrievedOn: "2026-11-07"},
		{URL: PR72025URL, Section: "LHDN Public Ruling No. 7/2025 (7th edition, 5 December 2025) " + ruling + "; read from the MAICSA mirror (docs/sources/pr7-2025) because the IRBM URL returned HTTP 500", RetrievedOn: "2026-11-16"},
	}
}
```

In `YA2025Draft`, replace the `Groups:` line with:

```go
		Groups: []Group{
			{ID: "LIFESTYLE", LimitSen: 250000},
			{ID: "MEDICAL_SELF_FAMILY", LimitSen: 1000000},
			// s.46(1)(h): RM1,000 within the RM10,000 of s.46(1)(g).
			{ID: "MEDICAL_SCREENING", Parent: "MEDICAL_SELF_FAMILY", LimitSen: 100000},
			{ID: "EDUCATION_SELF", LimitSen: 700000},
			// s.46(1)(c): one RM8,000 cap for all parents and grandparents together (interpretation; PR 7/2025 Example 3).
			{ID: "PARENTS", LimitSen: 800000},
			{ID: "DISABILITY_EQUIPMENT", LimitSen: 600000},
		},
```

After the `MEDICAL_DENTAL` rule, add:

```go
			// s.46(1)(c) parents and grandparents. No apportionment between siblings: each claimant claims what they paid.
			{ID: "PARENTS_MEDICAL", Group: "PARENTS", LimitSen: 800000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_medical_treatment", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c)(i) as amended by Finance Act 2024 s.9(a); RM8,000 is read as one cap for all parents and grandparents (interpretation)", "paragraphs 6.2.1-6.2.5, Example 3 and Appendix")},
			{ID: "PARENTS_CARER", Group: "PARENTS", LimitSen: 800000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_carer", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c)(ii) as amended by Finance Act 2024 s.9(a)", "paragraphs 6.2.1-6.2.4")},
			{ID: "PARENTS_CHECKUP", Group: "PARENTS", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "parent_resident_in_malaysia", "treatment_in_malaysia", "qualifying_examination", "no_other_claim"},
				Sources: medicalSources("s.46(1)(c) proviso (e) as amended by Finance Act 2024 s.9(a): complete medical examination including any vaccination, up to RM1,000", "paragraph 6.2.3(c) and Example 3")},
			// s.46(1)(d). One expense cannot also be claimed under (c): the shared expense_key blocks it (DUPLICATE_EXPENSE).
			{ID: "DISABILITY_EQUIPMENT", Group: "DISABILITY_EQUIPMENT", LimitSen: 600000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "disability_certified_by_jkm", "qualifying_supporting_equipment", "no_other_claim"},
				Sources: medicalSources("s.46(1)(d)", "paragraphs 3.6, 6.3.1-6.3.2 and Example 4")},
			// s.46(1)(e): a fixed RM7,000 for a disabled individual; claim with amount_sen 0.
			{ID: "DISABLED_SELF", FixedAmount: true, LimitSen: 700000, RequiredFacts: []string{"resident", "within_year", "disability_certified_by_jkm"},
				Sources: medicalSources("s.46(1)(e) as amended by Finance Act 2024 s.9(b) (RM6,000 to RM7,000 from YA 2025)", "paragraphs 3.6 and 6.4")},
			{ID: "MEDICAL_SERIOUS_DISEASE", Group: "MEDICAL_SELF_FAMILY", LimitSen: 1000000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "qualifying_serious_disease", "no_other_claim"},
				Sources: medicalSources("s.46(1)(g)(i); serious disease read as the diseases named in PR 7/2025 paragraph 6.6.2 (interpretation)", "paragraphs 6.6.1-6.6.2 and Examples 8, 11, 13, 16, 17, 19")},
			{ID: "MEDICAL_FERTILITY", Group: "MEDICAL_SELF_FAMILY", LimitSen: 1000000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "medical_practitioner_certified", "qualifying_fertility_treatment", "individual_married", "no_other_claim"},
				Sources: medicalSources("s.46(1)(g)(ii)", "paragraph 6.6.3 and Examples 9, 10, 11")},
			{ID: "MEDICAL_SCREENING_EXAM", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_examination", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(i) as amended by Finance Act 2024 s.9(c)", "paragraphs 6.7.1-6.7.4 and Example 14")},
			{ID: "MEDICAL_SCREENING_DETECTION_TEST", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_detection_test", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(i) as amended by Finance Act 2024 s.9(c)", "paragraph 6.7.6 and Examples 16, 17")},
			{ID: "MEDICAL_SCREENING_SELF_TEST_DEVICE", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_self_test_device", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(ii) as amended by Finance Act 2024 s.9(c): registered under the Medical Device Act 2012 (Act 737), not for business use", "paragraph 6.7.7 and Example 17")},
			{ID: "MEDICAL_SCREENING_MENTAL_HEALTH", Group: "MEDICAL_SCREENING", LimitSen: 100000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "qualifying_mental_health_provider", "no_other_claim"},
				Sources: medicalSources("s.46(1)(h)(iii)", "paragraph 6.7.5 and Example 14")},
			{ID: "MEDICAL_LEARNING_DISABILITY", Group: "MEDICAL_SELF_FAMILY", LimitSen: 600000, RequiredFacts: []string{"resident", "within_year", "qualifying_beneficiary", "paid_by_claimant", "treatment_in_malaysia", "child_age_18_or_below", "qualifying_learning_disability", "allied_health_practitioner_registered", "no_other_claim"},
				Sources: medicalSources("s.46(1)(ha) as amended by Finance Act 2024 s.9(d) (RM6,000 for YA 2025; RM10,000 starts YA 2026)", "paragraphs 6.8.1-6.8.3 and Examples 18, 19, 20")},
```

`MEDICAL_VACCINATION` and `MEDICAL_DENTAL` keep their current facts and sources; the recorded matrix `relief_cases.json` depends on them.

- [ ] **Step 6: Let fixed rules pass the draft-leak test**

In `engine_unified_test.go`, inside `TestDraftCatalogNeverProducesApprovedAmounts`, replace `c := claimFor(t, p, "a", rule.ID, 1)` with:

```go
		amt := int64(1)
		if rule.FixedAmount {
			amt = 0
		}
		c := claimFor(t, p, "a", rule.ID, amt)
```

- [ ] **Step 7: Run the policy tests**

Run: `cd services/api && go test ./internal/policy/ -count=1`
Expected: PASS, including `TestEveryRuleFactIsCanonicalAndEveryCanonicalFactIsUsed`, `TestUnifiedEngineMatchesRecordedEligibilityBehaviour` and all 17 `TestPR72025WorkedExamples` subtests.

- [ ] **Step 8: Commit**

```bash
git add services/api/internal/policy/facts.go services/api/internal/policy/catalog.go services/api/internal/policy/engine_unified_test.go services/api/internal/policy/medical_test.go services/api/internal/policy/testdata/pr7_2025_cases.json
git commit -m "feat(policy): YA2025 draft medical reliefs with PR 7/2025 golden fixtures"
```

---

### Task 3: Docs and registry notes

**Files:**
- Create: `docs/RELIEF_MEDICAL_YA2025.md`
- Modify: `docs/ROADMAP.md:22,41`, `docs/ASSESSMENT_PREVIEW.md:22-28`
- Modify: `services/api/internal/policy/registry.go:89-103` (notes only)

**Interfaces:**
- Consumes: rule IDs, group IDs, fact names and fixture outcomes defined in Task 2 (copied in the doc below). No Go API.
- Produces: docs only; `registry.go` notes keep amounts, kinds and the substring `Finance Act 2024` (pinned by `registry_act53_test.go`).

- [ ] **Step 1: Write `docs/RELIEF_MEDICAL_YA2025.md`**

Follow the layout of `docs/RELIEF_LIFESTYLE_YA2025.md`. Required sections and content:

1. Title `# Medical reliefs, YA 2025 — draft model` and the status line: `Status: **DRAFT. Not approved, not professionally reviewed, never reduces chargeable income.**`
2. `## Sources` — the spec's Sources table (Act 53 May 2024 text plus Finance Act 2024 s.9(a)–(d); Act 53 1 Jan 2026 text only for YA 2026 differences; PR 7/2025 with the IRBM URL, HTTP 500 on 2026-11-07 and 2026-11-16, the MAICSA mirror URL, mirror PDF SHA-256 `47e07419363bc88b0fbffc404a92fc625bf96c84a4f3878ee8f0e5da30e5d10b`, pointer to `docs/sources/pr7-2025/`; Nota Penerangan BE YA 2025 at `https://ef.hasil.gov.my/eBE2026/Pdf/Nota_BE.pdf`, accessed 2026-11-07, with the warning that `eBE2025/Pdf/Nota_BE_e.pdf` is the YA 2024 notes). State: the Act text is the AGC updated text, not an authentic text; the ruling is an administrative interpretation under s.138A.
3. `## What is modelled` — the spec's §1 rule table, plus a column "Required facts" listing each rule's `RequiredFacts` from Task 2 Step 5. Note `MEDICAL_SCREENING` is nested in `MEDICAL_SELF_FAMILY`.
4. `## Claim shape` — expense rules as today; `DISABLED_SELF` sends `amount_sen` 0 or omits it, with `expense_key` = DSW certificate number and `evidence_reference` = the certificate; any other amount is HTTP 400.
5. `## Interpretations for tax-professional review` — the spec's four numbered interpretations (§2), verbatim in substance.
6. `## Allocation order` — the spec's §4 known limit: claim-ID order decides the per-line split when a group cap binds; totals do not depend on it; LHDN prescribes no order (Example 19).
7. `## Worked examples covered` — table of example → expected lines and total, copied from the spec's Testing table (Examples 3, 4, 8, 9, 10, 11, 13, 14, 16, 17, 18, 19, 20), with the encodings from `pr7_2025_cases.json`. Then: "Not covered: Examples 12 and 15 (joint assessment, s.46(3) deeming); Examples 5–7 (disabled individual and spouse status, later batch)."
8. `## Not modelled` — disabled spouse and child (s.45A, s.47(1)(b), s.48(2)(b)); joint assessment; YA 2026 changes (Finance Act 2025 s.6: NPRA-registered vaccines, learning disability RM10,000); "other similar diseases" under (g)(i) (IRBM guideline not held).

- [ ] **Step 2: Update `docs/ROADMAP.md`**

After line 22 add: `- Medical reliefs (s.46(1)(c), (d), (e), (g), (h), (ha)) modelled as Draft rules with PR 7/2025 golden fixtures (`docs/RELIEF_MEDICAL_YA2025.md`).`
In line 41, change "Rules exist only for lifestyle, education, vaccination and dental, first-home interest" to "Rules exist only for lifestyle, education, the medical batch (parents, disability equipment, disabled individual, serious disease, fertility, vaccination, dental, screening, learning disability), first-home interest", and delete "parents' medical, disability equipment, disabled individual, serious disease and fertility, medical screening, learning disability, " from the "No rule yet for" list.

- [ ] **Step 3: Update `docs/ASSESSMENT_PREVIEW.md`**

After the claim example (line 25 block), add:

```markdown
A fixed-amount rule (currently `DISABLED_SELF`, RM7,000) takes `amount_sen` 0 or no `amount_sen`; the engine uses the rule limit. Any other amount returns HTTP 400:

    {"id":"c2","rule_id":"DISABLED_SELF","expense_key":"dsw-card-no","evidence_reference":"dsw-card","facts":{"resident":true,"within_year":true,"disability_certified_by_jkm":true}}

`GET /v1/policy/catalog` marks such rules `"fixed_amount": true`; groups may carry `"parent"` (for example `MEDICAL_SCREENING` inside `MEDICAL_SELF_FAMILY`).
```

- [ ] **Step 4: Cite PR 7/2025 in `registry.go` notes**

Append to the existing note text (keep every existing word, kind and the `Finance Act 2024` substring):
- `"PARENTS"`: ` PR 7/2025 Example 3 applies one RM8,000 cap to parents and grandparents together.`
- `"DISABLED_SELF"`: ` PR 7/2025 paragraphs 3.6 and 6.4: disabled person means certified by the Department of Social Welfare (DSW/JKM).`
- `"MEDICAL_SCREENING"`: ` PR 7/2025 paragraphs 6.7.1-6.7.7 and Examples 14, 16, 17; laboratory tests alone are not a complete examination (6.7.4).`
- `"MEDICAL_LEARNING_DISABILITY"`: ` PR 7/2025 paragraph 6.8 and Examples 18-20: child aged 18 or below, RM6,000 per taxpayer.`

`registry.go` has no eBE2025 URL (checked 2026-11-16 with `grep -rn eBE2025 services/api`), so no URL correction is needed.

- [ ] **Step 5: Run the registry tests**

Run: `cd services/api && go test ./internal/policy/ -run Registry -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add docs/RELIEF_MEDICAL_YA2025.md docs/ROADMAP.md docs/ASSESSMENT_PREVIEW.md services/api/internal/policy/registry.go
git commit -m "docs: YA2025 medical relief rules, interpretations and PR 7/2025 coverage"
```

---

### Task 4: Verification before merge

- [ ] **Step 1: Static checks**

Run: `cd services/api && gofmt -l . && go vet ./...`
Expected: no output from `gofmt -l`; `go vet` exits 0.

- [ ] **Step 2: Unit tests with race detector**

Run: `cd services/api && go mod tidy && git diff --exit-code go.mod go.sum && go test -race -count=1 ./...`
Expected: all packages `ok`.

- [ ] **Step 3: PostgreSQL integration tests on a throwaway PostgreSQL 16**

```bash
docker run -d --rm --name mytax-pg-verify -e POSTGRES_PASSWORD=ci-postgres-only -e POSTGRES_DB=mytax_ci -p 55432:5432 postgres:16
until docker exec mytax-pg-verify pg_isready -U postgres -d mytax_ci; do sleep 1; done
cd services/api
PGHOST=localhost PGPORT=55432 PGUSER=postgres PGPASSWORD=ci-postgres-only PGDATABASE=mytax_ci psql -X -v ON_ERROR_STOP=1 -f testdata/security_fixture.sql
MYTAX_TEST_DATABASE_URL='postgres://mytax_app:ci-only-insecure-password@localhost:55432/mytax_ci?sslmode=disable' \
MYTAX_TEST_TENANT_A=11111111-1111-4111-8111-111111111111 MYTAX_TEST_TENANT_B=22222222-2222-4222-8222-222222222222 \
go test -race -count=1 -v ./... 2>&1 | tee /tmp/mytax-go-test.log
grep -E -- '--- SKIP' /tmp/mytax-go-test.log
docker stop mytax-pg-verify
```

Expected: PostgreSQL RLS tests PASS. Any `--- SKIP` line other than the live Keycloak tests (which need the CI Keycloak) is a failure. Report the Keycloak skips as "not verified locally; CI covers them".

- [ ] **Step 4: Push and confirm CI**

Push `feat/medical-relief-batch` and open a PR. CI job "API and database security tests" must be green with zero skips. Report separately: committed, locally tested, CI passed, independently reviewed (not done: tax-professional review is required before any rule leaves Draft).
