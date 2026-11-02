package calculation

import "testing"

func TestDividendAlwaysNeedsReviewAndNeverComputes(t *testing.T) {
	for _, d := range []int64{0, 1, 9999999, 10000000, 10000001, 11000000, 500000000} {
		r, err := ReviewIndividualDividend(2025, d)
		if err != nil || r.Status != "REVIEW_REQUIRED" || r.Computed || len(r.Reasons) == 0 {
			t.Fatalf("dividend %d must need review: %+v %v", d, r, err)
		}
	}
}

func TestDividendIllustrationBoundaries(t *testing.T) {
	cases := []struct{ d, want int64 }{
		{0, 0}, {9999999, 0}, {10000000, 0}, {10000001, 0}, {10000050, 1}, {11000000, 20000}, {20000000, 200000},
	}
	for _, tc := range cases {
		r, _ := ReviewIndividualDividend(2025, tc.d)
		if r.IllustrativeExcessTaxSen != tc.want {
			t.Fatalf("d=%d illustration=%d want=%d", tc.d, r.IllustrativeExcessTaxSen, tc.want)
		}
	}
}

func TestDividendReasonsByThreshold(t *testing.T) {
	has := func(r DividendReview, code string) bool {
		for _, x := range r.Reasons {
			if x == code {
				return true
			}
		}
		return false
	}
	low, _ := ReviewIndividualDividend(2025, 10000000)
	if !has(low, "EXEMPTION_CONDITION_UNVERIFIED_COMPANY_NOT_ENTITLED_TO_DEDUCT_TAX") || has(low, "EXCESS_ONLY_OR_WHOLE_DIVIDEND_UNRESOLVED") {
		t.Fatalf("%+v", low)
	}
	high, _ := ReviewIndividualDividend(2025, 10000001)
	for _, c := range []string{"AGGREGATION_BASIS_UNRESOLVED_PER_COMPANY_OR_TOTAL", "EXCESS_ONLY_OR_WHOLE_DIVIDEND_UNRESOLVED", "CHARGEABLE_INCOME_PRESCRIPTION_NOT_AVAILABLE"} {
		if !has(high, c) {
			t.Fatalf("missing %s: %+v", c, high)
		}
	}
}

func TestDividendRejectsBadYearAndNegative(t *testing.T) {
	for _, ya := range []int{2024, 2026, 0} {
		if _, err := ReviewIndividualDividend(ya, 1); err == nil {
			t.Fatalf("year %d accepted", ya)
		}
	}
	if _, err := ReviewIndividualDividend(2025, -1); err == nil {
		t.Fatal("negative accepted")
	}
}

func TestDividendIllustrationRoundsHalfUpToTheSen(t *testing.T) {
	// 2% of an excess of 24 sen is 0.48 sen, of 25 sen is 0.50 sen (rounds up), of 74 sen is 1.48 sen.
	cases := []struct{ excess, want int64 }{{24, 0}, {25, 1}, {74, 1}, {75, 2}}
	for _, tc := range cases {
		r, _ := ReviewIndividualDividend(2025, DividendThresholdSen+tc.excess)
		if r.IllustrativeExcessTaxSen != tc.want {
			t.Fatalf("excess=%d got %d want %d", tc.excess, r.IllustrativeExcessTaxSen, tc.want)
		}
	}
}
