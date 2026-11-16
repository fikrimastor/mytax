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
//
// Allocation order: because claims are allocated in claim-ID order, when two claims
// compete for a cap or for the same exclusive family, the lower claim ID wins.
//
// Nested groups: a claim is limited by its rule, its group and every ancestor group, and its
// allowed amount is charged to all of them. When a group cap binds, the per-line split
// depends on claim-ID order; the group total does not.
//
// Fixed amounts: a claim on a FixedAmount rule sends amount_sen 0 and is granted LimitSen.
//
// Expense keys: an ExpenseKey is registered even when its claim is blocked for another
// reason, so a later claim reusing that key is flagged as a duplicate (fails closed).
//
// Package status: Evaluate trusts Package.Status. A caller that wants Approved amounts
// must first run ValidateRelease.
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
	rules, groups, err := validateStructure(p)
	if err != nil {
		return Result{}, err
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
		if seenID[c.ID] {
			return Result{}, errors.New("duplicate claim ID")
		}
		seenID[c.ID] = true
		d := Decision{ClaimID: c.ID, Status: "REVIEW_REQUIRED", Reasons: []string{}}
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
			if allowed < claimed {
				d.Reasons = append(d.Reasons, "LIMIT_REACHED")
			}
		}
		res.Decisions = append(res.Decisions, d)
	}
	return res, nil
}
