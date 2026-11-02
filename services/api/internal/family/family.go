// Package family computes CANDIDATE spouse and child deductions under the Income Tax Act 1967
// s.45A, s.47 and s.48, for YA2025 only. Sources: Act 53 updated texts (21 May 2024 and
// 1 January 2026; not authentic texts) and Finance Act 2024 (Act 862) s.8, s.11 and s.12.
//
// Everything is a draft candidate. No result is approved, reduces chargeable income, or is
// filing-ready. All facts are caller-asserted and unverified. Where the supplied texts leave a
// point open, the result is REVIEW_REQUIRED with a zero candidate amount (fail closed).
package family

import (
	"errors"
	"math"
	"sort"
)

const (
	YA = 2025

	wifeBaseSen            = int64(400000) // s.47(1)(a); s.45A(1) husband base
	disabledSpouseExtraSen = int64(600000) // s.47(1)(b) and s.45A(1), as amended by Finance Act 2024 for YA2025
	childBaseSen           = int64(200000) // s.48(2)(a)
	childDisabledSen       = int64(800000) // s.48(2)(b), as amended by Finance Act 2024 for YA2025
	childUpliftMultiple    = int64(4)      // s.48(3)(a)(i)
	childDisabledExtraSen  = int64(800000) // s.48(3)(a)(ii)

	CatUnder18     = "UNDER_18"               // s.48(1)(a)
	CatInstruction = "FULL_TIME_INSTRUCTION"  // s.48(1)(b)
	CatArticles    = "ARTICLES_OR_INDENTURES" // s.48(1)(c)
	CatDisabled    = "DISABLED"               // s.48(1)(d)

	RoleHusband = "HUSBAND" // claims for a wife under s.47
	RoleWife    = "WIFE"    // claims for a husband under s.45A

	stateReview = "REVIEW_REQUIRED"
	reasonDraft = "DRAFT_LEGAL_RULE_NOT_APPROVED"
)

type Decision struct {
	ID           string   `json:"id"`
	State        string   `json:"state"`
	CandidateSen int64    `json:"candidate_sen"`
	Reasons      []string `json:"reasons"`
}

type Result struct {
	AssessmentYear    int        `json:"assessment_year"`
	Status            string     `json:"status"`
	Decisions         []Decision `json:"decisions"`
	CandidateTotalSen int64      `json:"candidate_total_sen"`
	ApprovedTotalSen  int64      `json:"approved_total_sen"`
	FilingReady       bool       `json:"filing_ready"`
	Notice            string     `json:"notice"`
}

func newResult() Result {
	return Result{AssessmentYear: YA, Status: "DRAFT_NOT_FOR_FILING", Decisions: []Decision{}, FilingReady: false,
		Notice: "Draft candidates from unverified facts. Not approved, not filing-ready, and must not reduce chargeable income."}
}

// ---------------------------------------------------------------- children (s.48)

type Child struct {
	ID        string
	Category  string
	Unmarried bool
	// Category facts, caller-asserted.
	UnderEighteenAtAnyTime      bool // CatUnder18
	FullTimeInstruction         bool // CatInstruction
	ServingArticlesOrIndentures bool // CatArticles
	DisabilityProvenToDG        bool // CatDisabled: "proved to the satisfaction of the Director General"
	// s.48(3): child over 18 in higher education or serving articles/indentures.
	Over18                    bool
	HigherEducationOrArticles bool
	ApprovedByAuthority       bool // s.48(3)(b)
	OutsideMalaysia           bool
	DegreeOrEquivalent        bool // required outside Malaysia
	// s.48(5)
	TotalIncomeSen      int64
	TotalIncomeVerified bool
}

type ChildClaim struct {
	ID               string
	ClaimantID       string
	ChildID          string
	ClaimantResident bool
	PaidMaintenance  bool // pays wholly or in part for maintenance (or the instruction)
	// s.48(1) proviso: a wife living with her husband and assessed separately may elect in
	// writing that the whole deduction is allowed to her.
	WifeElectsWhole                         bool
	WifeLivingWithHusbandAssessedSeparately bool
}

// wholeAmount is the deduction for one claimant before s.48(4) sharing and s.48(5).
func wholeAmount(c Child) (int64, []string) {
	var reasons []string
	need := func(ok bool, fact string) {
		if !ok {
			reasons = append(reasons, "FACT_NOT_VERIFIED:"+fact)
		}
	}
	need(c.Unmarried, "unmarried")
	var base int64
	switch c.Category {
	case CatUnder18:
		need(c.UnderEighteenAtAnyTime, "under_eighteen_at_any_time")
		base = childBaseSen
	case CatInstruction:
		need(c.FullTimeInstruction, "full_time_instruction")
		base = childBaseSen
	case CatArticles:
		need(c.ServingArticlesOrIndentures, "serving_articles_or_indentures")
		base = childBaseSen
	case CatDisabled:
		need(c.DisabilityProvenToDG, "disability_proven_to_director_general")
		base = childDisabledSen
	default:
		return 0, []string{"UNKNOWN_CHILD_CATEGORY"}
	}
	if c.Over18 && c.HigherEducationOrArticles {
		if c.Category == CatUnder18 {
			reasons = append(reasons, "UNDER_18_CATEGORY_CONFLICTS_WITH_OVER_18")
		}
		need(c.ApprovedByAuthority, "institution_approved_by_government_authority")
		if c.OutsideMalaysia {
			need(c.DegreeOrEquivalent, "degree_or_equivalent_outside_malaysia")
		}
		switch c.Category {
		case CatInstruction, CatArticles:
			base *= childUpliftMultiple // s.48(3)(a)(i): four times, in substitution
		case CatDisabled:
			base += childDisabledExtraSen // s.48(3)(a)(ii): a further RM8,000
		}
	}
	if len(reasons) > 0 {
		return 0, reasons
	}
	return base, nil
}

// ChildDeductions computes candidate s.48 deductions. Claims are processed in claim ID order.
func ChildDeductions(children []Child, claims []ChildClaim) (Result, error) {
	if len(children) > 200 || len(claims) > 1000 {
		return Result{}, errors.New("too many records")
	}
	kids := map[string]Child{}
	for _, c := range children {
		if c.ID == "" || c.TotalIncomeSen < 0 {
			return Result{}, errors.New("invalid child")
		}
		if _, dup := kids[c.ID]; dup {
			return Result{}, errors.New("duplicate child ID")
		}
		kids[c.ID] = c
	}
	ordered := append([]ChildClaim(nil), claims...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	seen := map[string]bool{}
	for _, c := range ordered {
		if c.ID == "" || c.ClaimantID == "" || seen[c.ID] {
			return Result{}, errors.New("invalid or duplicate claim ID")
		}
		seen[c.ID] = true
	}

	out := newResult()
	decisions := map[string]*Decision{}
	for _, c := range ordered {
		decisions[c.ID] = &Decision{ID: c.ID, State: stateReview, Reasons: []string{}}
	}
	byChild := map[string][]ChildClaim{}
	pair := map[string]bool{}
	for _, c := range ordered {
		d := decisions[c.ID]
		child, ok := kids[c.ChildID]
		switch {
		case !ok:
			d.Reasons = append(d.Reasons, "UNKNOWN_CHILD")
		case pair[c.ClaimantID+"\x00"+c.ChildID]:
			d.Reasons = append(d.Reasons, "DUPLICATE_CLAIM_FOR_SAME_CHILD_BY_SAME_CLAIMANT")
		default:
			pair[c.ClaimantID+"\x00"+c.ChildID] = true
			if !c.ClaimantResident {
				d.Reasons = append(d.Reasons, "FACT_NOT_VERIFIED:claimant_resident")
			}
			if !c.PaidMaintenance {
				d.Reasons = append(d.Reasons, "FACT_NOT_VERIFIED:paid_maintenance_or_instruction")
			}
			if _, why := wholeAmount(child); len(why) > 0 {
				d.Reasons = append(d.Reasons, why...)
			}
			if len(d.Reasons) == 0 {
				byChild[c.ChildID] = append(byChild[c.ChildID], c)
			}
		}
	}

	childIDs := make([]string, 0, len(byChild))
	for id := range byChild {
		childIDs = append(childIDs, id)
	}
	sort.Strings(childIDs)
	for _, id := range childIDs {
		child := kids[id]
		whole, _ := wholeAmount(child)
		group := byChild[id]
		share := make(map[string]int64, len(group))
		block := func(reason string) {
			for _, c := range group {
				decisions[c.ID].Reasons = append(decisions[c.ID].Reasons, reason)
			}
		}
		electing := 0
		for _, c := range group {
			if c.WifeElectsWhole {
				electing++
			}
		}
		switch {
		case len(group) > 2:
			block("MORE_THAN_TWO_CLAIMANTS_UNRESOLVED")
			continue
		case electing > 1:
			block("ELECTION_AMBIGUOUS")
			continue
		case electing == 1 && len(group) == 2:
			valid := false
			for _, c := range group {
				if c.WifeElectsWhole && c.WifeLivingWithHusbandAssessedSeparately {
					valid = true
				}
			}
			if !valid {
				block("ELECTION_CONDITION_NOT_VERIFIED")
				continue
			}
			for _, c := range group {
				if c.WifeElectsWhole {
					share[c.ID] = whole
				} else {
					decisions[c.ID].Reasons = append(decisions[c.ID].Reasons, "WIFE_ELECTION_GIVES_WHOLE_DEDUCTION_TO_OTHER_CLAIMANT")
				}
			}
		case electing == 1:
			block("ELECTION_WITHOUT_TWO_CLAIMANTS")
			continue
		case len(group) == 2:
			for _, c := range group {
				share[c.ID] = whole / 2 // s.48(4): fifty per cent each
			}
		default:
			share[group[0].ID] = whole
		}
		for _, c := range group {
			s, ok := share[c.ID]
			if !ok {
				continue
			}
			d := decisions[c.ID]
			switch {
			case !child.TotalIncomeVerified:
				d.Reasons = append(d.Reasons, "CHILD_TOTAL_INCOME_NOT_VERIFIED")
			case child.TotalIncomeSen > whole:
				d.Reasons = append(d.Reasons, "CHILD_INCOME_EXCEEDS_DEDUCTION")
			case child.TotalIncomeSen > s:
				// s.48(5) compares with "the deduction that would otherwise be allowed to that
				// individual". Whether that is before or after the s.48(4) halving is not settled.
				d.Reasons = append(d.Reasons, "CHILD_INCOME_TEST_BASIS_UNRESOLVED")
			default:
				d.CandidateSen = s
			}
		}
	}

	for _, c := range ordered {
		d := decisions[c.ID]
		if d.CandidateSen > 0 {
			if out.CandidateTotalSen > math.MaxInt64-d.CandidateSen {
				return Result{}, errors.New("overflow")
			}
			out.CandidateTotalSen += d.CandidateSen
			d.Reasons = append(d.Reasons, reasonDraft)
		}
		out.Decisions = append(out.Decisions, *d)
	}
	return out, nil
}

// ---------------------------------------------------------------- spouse (s.47 and s.45A)

type SpouseInput struct {
	ID               string
	Role             string
	ClaimantResident bool
	LivingTogether   bool // spouse lived with the claimant in the basis year
	// s.47(4): a wife assessed separately in her own name gives no s.47(1) deduction.
	SpouseAssessedSeparately bool
	SpouseDisabled           bool
	DisabilityProven         bool
	// s.47(6) / s.45A(2): gross foreign-sourced income of the spouse.
	SpouseForeignIncomeSen  int64
	ForeignIncomeVerified   bool
	AlimonyPaidSen          int64 // s.47(2), HUSBAND role only
	S45AConditionMet        bool  // WIFE role: s.45A(1)(a), (b) or (c)
	OnlyOneWifeClaimingS45A bool  // WIFE role: s.45A(1) proviso
}

// SpouseDeduction computes one candidate s.47 (husband) or s.45A (wife) deduction.
func SpouseDeduction(in SpouseInput) (Result, error) {
	if in.Role != RoleHusband && in.Role != RoleWife {
		return Result{}, errors.New("unknown role")
	}
	if in.ID == "" || in.SpouseForeignIncomeSen < 0 || in.AlimonyPaidSen < 0 {
		return Result{}, errors.New("invalid input")
	}
	out := newResult()
	d := Decision{ID: in.ID, State: stateReview, Reasons: []string{}}
	need := func(ok bool, fact string) {
		if !ok {
			d.Reasons = append(d.Reasons, "FACT_NOT_VERIFIED:"+fact)
		}
	}
	need(in.ClaimantResident, "claimant_resident")

	var base int64 // s.47(1)(a) or s.45A(1) base
	var alimony int64
	if in.Role == RoleHusband {
		if in.AlimonyPaidSen > 0 {
			alimony = in.AlimonyPaidSen
		}
		if in.LivingTogether && !in.SpouseAssessedSeparately {
			base = wifeBaseSen
		} else if in.LivingTogether {
			d.Reasons = append(d.Reasons, "S47_4_WIFE_ASSESSED_SEPARATELY_NO_S47_1_DEDUCTION")
		} else if alimony == 0 {
			d.Reasons = append(d.Reasons, "FACT_NOT_VERIFIED:wife_living_together")
		}
	} else {
		need(in.S45AConditionMet, "s45a_condition_met")
		need(in.OnlyOneWifeClaimingS45A, "only_one_wife_claiming")
		if in.S45AConditionMet && in.OnlyOneWifeClaimingS45A {
			base = wifeBaseSen
		}
	}

	// Foreign income test: s.47(6) (non-disabled wife) and s.45A(2) (non-disabled husband), "more than" RM4,000.
	if base > 0 && !in.SpouseDisabled {
		if !in.ForeignIncomeVerified {
			d.Reasons = append(d.Reasons, "SPOUSE_FOREIGN_INCOME_NOT_VERIFIED")
			base = 0
		} else if in.SpouseForeignIncomeSen > wifeBaseSen {
			d.Reasons = append(d.Reasons, "SPOUSE_FOREIGN_INCOME_EXCEEDS_DEDUCTION")
			base = 0
		}
	}

	total := base
	if in.Role == RoleHusband {
		// s.47(3): the deduction under (1)(a) plus alimony under (2) must not exceed RM4,000.
		if total > math.MaxInt64-alimony {
			return Result{}, errors.New("overflow")
		}
		total += alimony
		if total > wifeBaseSen {
			total = wifeBaseSen
		}
	}
	if in.SpouseDisabled {
		if in.DisabilityProven {
			if base > 0 {
				total += disabledSpouseExtraSen // the further amount sits under s.47(1)(b) / s.45A(1), so it needs the base conditions
			}
		} else {
			d.Reasons = append(d.Reasons, "FACT_NOT_VERIFIED:spouse_disability_proven")
		}
	}
	if !in.ClaimantResident {
		total = 0
	}
	d.CandidateSen = total
	if total > 0 {
		d.Reasons = append(d.Reasons, reasonDraft)
		out.CandidateTotalSen = total
	}
	out.Decisions = append(out.Decisions, d)
	return out, nil
}
