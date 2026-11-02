package calculation

import "testing"

func TestRebateBoundaryAtRM35000(t *testing.T) {
	cases := []struct {
		income, want int64
	}{{0, 40000}, {3499999, 40000}, {3500000, 40000}, {3500001, 0}}
	for _, tc := range cases {
		r, err := ProvisionalRebates(100000, tc.income, RebateFacts{SelfDeductionClaimed: true})
		if err != nil || r.ProvisionalRebateSen != tc.want {
			t.Fatalf("income=%d rebate=%d want=%d err=%v", tc.income, r.ProvisionalRebateSen, tc.want, err)
		}
	}
}

func TestRebateSelfSpouseAndWifeAreSeparateRM400(t *testing.T) {
	r, err := ProvisionalRebates(500000, 3000000, RebateFacts{SelfDeductionClaimed: true, SpouseDeductionClaimed: true, WifeHusbandDeductionClaimed: true})
	if err != nil || r.UncappedProvisionalRebateSen != 120000 || r.TaxAfterProvisionalRebateSen != 380000 {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestRebateNeverCreatesRefundAndLosesExcess(t *testing.T) {
	r, err := ProvisionalRebates(10000, 3000000, RebateFacts{SelfDeductionClaimed: true, SpouseDeductionClaimed: true, ZakatPaidSen: 500000, ZakatReceiptFromAuthority: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.ProvisionalRebateSen != 10000 || r.TaxAfterProvisionalRebateSen != 0 || r.RefundSen != 0 || r.LostExcessRebateSen != r.UncappedProvisionalRebateSen-10000 || r.LostExcessRebateSen <= 0 {
		t.Fatalf("refund or carry-forward leaked: %+v", r)
	}
	z, _ := ProvisionalRebates(0, 0, RebateFacts{SelfDeductionClaimed: true})
	if z.TaxAfterProvisionalRebateSen != 0 || z.ProvisionalRebateSen != 0 || z.RefundSen != 0 {
		t.Fatalf("zero tax must stay zero: %+v", z)
	}
}

func TestZakatNeedsReceiptFromAuthority(t *testing.T) {
	r, _ := ProvisionalRebates(100000, 5000000, RebateFacts{ZakatPaidSen: 30000})
	if r.ProvisionalRebateSen != 0 || r.Lines[0].Reasons[0] != "ZAKAT_RECEIPT_FROM_STATUTORY_AUTHORITY_NOT_VERIFIED" {
		t.Fatalf("unverified zakat reduced tax: %+v", r)
	}
	r, _ = ProvisionalRebates(100000, 5000000, RebateFacts{ZakatPaidSen: 30000, ZakatReceiptFromAuthority: true})
	if r.ProvisionalRebateSen != 30000 || r.TaxAfterProvisionalRebateSen != 70000 {
		t.Fatalf("%+v", r)
	}
	// Zakat has no RM35,000 income test.
	r, _ = ProvisionalRebates(100000, 90000000, RebateFacts{ZakatPaidSen: 1, ZakatReceiptFromAuthority: true})
	if r.ProvisionalRebateSen != 1 {
		t.Fatalf("zakat wrongly income-tested: %+v", r)
	}
}

func TestUnclaimedRebatesAreNotApplied(t *testing.T) {
	r, err := ProvisionalRebates(50000, 1000000, RebateFacts{})
	if err != nil || r.ProvisionalRebateSen != 0 || len(r.Lines) != 0 || r.TaxAfterProvisionalRebateSen != 50000 {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestRebateRejectsNegativeAndStaysProvisional(t *testing.T) {
	for _, args := range [][3]int64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}} {
		if _, err := ProvisionalRebates(args[0], args[1], RebateFacts{ZakatPaidSen: args[2]}); err == nil {
			t.Fatalf("negative accepted: %v", args)
		}
	}
	r, _ := ProvisionalRebates(1, 1, RebateFacts{SelfDeductionClaimed: true})
	if r.Status != "PROVISIONAL_NOT_FILING_READY" {
		t.Fatalf("status %s", r.Status)
	}
}

func TestRebateOnRM30000IncomeFromRealGrossTax(t *testing.T) {
	g, err := ResidentGrossTax(2025, 3000000) // RM30,000 -> RM450
	if err != nil || g.GrossTaxSen != 45000 {
		t.Fatalf("%+v %v", g, err)
	}
	r, err := ProvisionalRebates(g.GrossTaxSen, g.ChargeableIncomeSen, RebateFacts{SelfDeductionClaimed: true})
	if err != nil || r.TaxAfterProvisionalRebateSen != 5000 {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestRebateCapBoundaryAtExactTax(t *testing.T) {
	facts := RebateFacts{SelfDeductionClaimed: true} // RM400 = 40000 sen
	cases := []struct{ tax, applied, lost, after int64 }{
		{39999, 39999, 1, 0}, {40000, 40000, 0, 0}, {40001, 40000, 0, 1},
	}
	for _, tc := range cases {
		r, err := ProvisionalRebates(tc.tax, 1000000, facts)
		if err != nil || r.ProvisionalRebateSen != tc.applied || r.LostExcessRebateSen != tc.lost || r.TaxAfterProvisionalRebateSen != tc.after || r.RefundSen != 0 {
			t.Fatalf("tax=%d: %+v %v", tc.tax, r, err)
		}
	}
}
