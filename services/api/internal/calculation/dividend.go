package calculation

import "errors"

// Individual dividend income above RM100,000 (Finance Act 2024, Act 862).
//
// Schedule 1 Part XXII (YA2025 onward, with s.6(1)(r)) charges, "notwithstanding Part I", an
// individual's dividend deemed derived from Malaysia under s.14 "in excess of one hundred
// thousand ringgit at the rate of 2 per cent on every ringgit of the chargeable income in
// respect of such dividend". Schedule 6 paragraph 12B(2) exempts a dividend to an individual
// amounting to RM100,000 or less where the paying company is not entitled to deduct tax.
//
// The supplied texts leave these open, so this package does NOT compute a tax liability:
//   - whether only the excess, or the whole dividend, is charged once the total exceeds RM100,000;
//   - whether RM100,000 is per company or per individual across all companies;
//   - the chargeable income "as prescribed by the Minister" where the individual has other income.
//
// The only number returned is an illustration of 2% on the excess, labelled as such.
const DividendThresholdSen = int64(10000000) // RM100,000

type DividendReview struct {
	AssessmentYear           int      `json:"assessment_year"`
	DividendSen              int64    `json:"dividend_sen"`
	Status                   string   `json:"status"`
	Reasons                  []string `json:"reasons"`
	IllustrativeExcessTaxSen int64    `json:"illustrative_excess_tax_sen"`
	Computed                 bool     `json:"computed"`
	Notice                   string   `json:"notice"`
}

func ReviewIndividualDividend(ya int, dividendSen int64) (DividendReview, error) {
	if ya != 2025 {
		return DividendReview{}, errors.New("dividend treatment not verified for requested assessment year")
	}
	if dividendSen < 0 {
		return DividendReview{}, errors.New("dividend must not be negative")
	}
	r := DividendReview{
		AssessmentYear: ya, DividendSen: dividendSen, Status: "REVIEW_REQUIRED", Computed: false,
		Reasons: []string{"AGGREGATION_BASIS_UNRESOLVED_PER_COMPANY_OR_TOTAL"},
		Notice:  "Not a tax computation. Finance Act 2024 sets a 2% charge on individual dividends above RM100,000 from YA2025; how it applies is not resolved by the supplied texts. A qualified person must decide.",
	}
	if dividendSen <= DividendThresholdSen {
		r.Reasons = append(r.Reasons, "EXEMPTION_CONDITION_UNVERIFIED_COMPANY_NOT_ENTITLED_TO_DEDUCT_TAX")
		return r, nil
	}
	excess := dividendSen - DividendThresholdSen
	r.Reasons = append(r.Reasons, "EXCESS_ONLY_OR_WHOLE_DIVIDEND_UNRESOLVED", "CHARGEABLE_INCOME_PRESCRIPTION_NOT_AVAILABLE")
	// 2% = excess/50 rounded half-up. Divide before multiplying so a near-MaxInt64 dividend
	// cannot overflow int64. The remainder r < 50, so r*2 < 100 never overflows either.
	// 2% of the remainder is r/50; half-up means anything at or above 25 rounds up.
	r.IllustrativeExcessTaxSen = excess / 50
	if excess%50 >= 25 {
		r.IllustrativeExcessTaxSen++
	}
	return r, nil
}
