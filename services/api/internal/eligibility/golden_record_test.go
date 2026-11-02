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
