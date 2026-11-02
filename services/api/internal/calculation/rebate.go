package calculation

import (
	"errors"
	"math"
)

// Provisional section 6A rebates for a resident individual (Income Tax Act 1967,
// updated texts as at 21 May 2024 and 1 January 2026; s.6A last amended by Act 823, YA2019).
//
// Everything here is PROVISIONAL and NOT filing-ready. The caller's flags are not proof:
// "was allowed a deduction" and "receipt from a statutory religious authority" are facts that
// this package cannot verify. A rebate never creates a refund (s.6A(4)): any excess over the
// tax charged is lost, not paid and not carried forward.
//
// Not implemented: the umrah departure-levy rebate (s.6A(2A)-(2B)), which needs a count of
// earlier rebated trips across years; and the rule that where Part XIV or XV of Schedule 1
// applies, the RM35,000 test uses chargeable income from all sources.
const (
	RebateThresholdSen = int64(3500000) // RM35,000: chargeable income must not exceed this (s.6A(2))
	RebateEachSen      = int64(40000)   // RM400 (s.6A(2)(a), (b), (c))
)

type RebateFacts struct {
	// Caller-asserted, unverified: the individual was allowed a deduction under s.46(1)(a).
	SelfDeductionClaimed bool `json:"self_deduction_claimed"`
	// Caller-asserted, unverified: a deduction under s.47(1) or (2) was allowed.
	SpouseDeductionClaimed bool `json:"spouse_deduction_claimed"`
	// Caller-asserted, unverified: a wife was allowed a deduction under s.45A.
	WifeHusbandDeductionClaimed bool `json:"wife_husband_deduction_claimed"`
	// Zakat, fitrah or other obligatory Islamic dues paid in the basis year (s.6A(3)).
	ZakatPaidSen int64 `json:"zakat_paid_sen"`
	// Evidence is a receipt from an appropriate religious authority established under written law.
	ZakatReceiptFromAuthority bool `json:"zakat_receipt_from_authority"`
}

type RebateLine struct {
	ID        string   `json:"id"`
	AmountSen int64    `json:"amount_sen"`
	Reasons   []string `json:"reasons"`
}

type RebateResult struct {
	GrossTaxSen                  int64        `json:"gross_tax_sen"`
	ChargeableIncomeSen          int64        `json:"chargeable_income_sen"`
	Lines                        []RebateLine `json:"lines"`
	UncappedProvisionalRebateSen int64        `json:"uncapped_provisional_rebate_sen"`
	ProvisionalRebateSen         int64        `json:"provisional_rebate_sen"`
	LostExcessRebateSen          int64        `json:"lost_excess_rebate_sen"`
	TaxAfterProvisionalRebateSen int64        `json:"tax_after_provisional_rebate_sen"`
	RefundSen                    int64        `json:"refund_sen"`
	Status                       string       `json:"status"`
	Notice                       string       `json:"notice"`
}

// ProvisionalRebates applies the s.6A rebates to an already computed gross tax. It never
// returns a negative tax or a refund.
func ProvisionalRebates(grossTaxSen, chargeableIncomeSen int64, f RebateFacts) (RebateResult, error) {
	if grossTaxSen < 0 || chargeableIncomeSen < 0 || f.ZakatPaidSen < 0 {
		return RebateResult{}, errors.New("amounts must not be negative")
	}
	out := RebateResult{
		GrossTaxSen: grossTaxSen, ChargeableIncomeSen: chargeableIncomeSen, Lines: []RebateLine{},
		Status: "PROVISIONAL_NOT_FILING_READY",
		Notice: "Provisional s.6A rebates from unverified caller facts. Not a filing-ready result. A rebate never gives a refund; any excess over the tax charged is lost (s.6A(4)). Umrah departure-levy rebate not computed.",
	}
	add := func(id string, amount int64, reasons ...string) error {
		if reasons == nil {
			reasons = []string{}
		}
		if amount > 0 && out.UncappedProvisionalRebateSen > math.MaxInt64-amount {
			return errors.New("rebate overflow")
		}
		out.Lines = append(out.Lines, RebateLine{ID: id, AmountSen: amount, Reasons: reasons})
		out.UncappedProvisionalRebateSen += amount
		return nil
	}
	fixed := func(id string, claimed bool) error {
		if !claimed {
			return nil
		}
		if chargeableIncomeSen > RebateThresholdSen {
			return add(id, 0, "CHARGEABLE_INCOME_ABOVE_RM35000")
		}
		return add(id, RebateEachSen, "DEDUCTION_ALLOWED_NOT_VERIFIED")
	}
	if err := fixed("REBATE_SELF_46_1_A", f.SelfDeductionClaimed); err != nil {
		return RebateResult{}, err
	}
	if err := fixed("REBATE_SPOUSE_47", f.SpouseDeductionClaimed); err != nil {
		return RebateResult{}, err
	}
	if err := fixed("REBATE_45A", f.WifeHusbandDeductionClaimed); err != nil {
		return RebateResult{}, err
	}
	if f.ZakatPaidSen > 0 {
		if !f.ZakatReceiptFromAuthority {
			if err := add("REBATE_ZAKAT", 0, "ZAKAT_RECEIPT_FROM_STATUTORY_AUTHORITY_NOT_VERIFIED"); err != nil {
				return RebateResult{}, err
			}
		} else if err := add("REBATE_ZAKAT", f.ZakatPaidSen, "RECEIPT_AUTHENTICITY_NOT_VERIFIED"); err != nil {
			return RebateResult{}, err
		}
	}
	out.ProvisionalRebateSen = out.UncappedProvisionalRebateSen
	if out.ProvisionalRebateSen > grossTaxSen {
		out.ProvisionalRebateSen = grossTaxSen
	}
	out.LostExcessRebateSen = out.UncappedProvisionalRebateSen - out.ProvisionalRebateSen
	out.TaxAfterProvisionalRebateSen = grossTaxSen - out.ProvisionalRebateSen
	return out, nil
}
