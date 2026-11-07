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

// Note: facts here are built from the NEW catalog, so this test cannot detect
// fact-list changes made when the engines were merged (education and dental
// now use the policy fact lists). TestMergedRulesUsePolicyFactSets pins those.
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
