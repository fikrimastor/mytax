// Package reconciliation supports an illustrative Form BE-style worksheet.
// All income and relief figures are user-supplied, not legally validated.
package reconciliation

import (
 "errors"
 "math"
 "github.com/fikrimastor/mytax/services/api/internal/calculation"
)
type Request struct {
 AssessmentYear int `json:"assessment_year"`
 DeclaredTotalIncomeSen int64 `json:"declared_total_income_sen"`
 DraftReliefsSen int64 `json:"draft_reliefs_sen"`
 VerifiedReliefsSen int64 `json:"verified_reliefs_sen"`
 EnteredRebatesSen int64 `json:"entered_rebates_sen"`
 PCBSen int64 `json:"pcb_sen"`
}
type Response struct {
 Status string `json:"status"`
 FilingReady bool `json:"filing_ready"`
 TotalIncomeSen int64 `json:"total_income_sen"`
 DraftReliefsExcludedSen int64 `json:"draft_reliefs_excluded_sen"`
 VerifiedReliefsAppliedSen int64 `json:"verified_reliefs_applied_sen"`
 IllustrativeChargeableIncomeSen int64 `json:"illustrative_chargeable_income_sen"`
 GrossTaxSen int64 `json:"gross_tax_sen"`
 EnteredRebatesSen int64 `json:"entered_rebates_sen"`
 TaxAfterEnteredRebatesSen int64 `json:"tax_after_entered_rebates_sen"`
 PCBSen int64 `json:"pcb_sen"`
 IllustrativeBalanceSen int64 `json:"illustrative_balance_sen"`
 Brackets []calculation.Line `json:"brackets"`
 Warnings []string `json:"warnings"`
}
func Preview(r Request)(Response,error){
 if r.AssessmentYear!=2025{return Response{},errors.New("only YA 2025 supported")}
 if r.DeclaredTotalIncomeSen<0 || r.DraftReliefsSen<0 || r.VerifiedReliefsSen<0 || r.EnteredRebatesSen<0 || r.PCBSen<0{return Response{},errors.New("amounts cannot be negative")}
 // No approved reliefs available from the policy system: reject any attempt to
 // mark a deduction as verified via a client-supplied number.
 if r.VerifiedReliefsSen!=0{return Response{},errors.New("verified reliefs cannot be supplied before policy approval")}
 if r.DeclaredTotalIncomeSen>math.MaxInt64/2 || r.PCBSen>math.MaxInt64/2{return Response{},errors.New("amount exceeds preview range")}
 // Rebate eligibility remains unimplemented. Reject claimed rebates rather
 // than silently validating a rebate based solely on income threshold.
 if r.EnteredRebatesSen!=0{return Response{},errors.New("rebate eligibility is not yet implemented")}
 tax,err:=calculation.ResidentGrossTax(2025,r.DeclaredTotalIncomeSen)
 if err!=nil{return Response{},err}
 return Response{
 Status:"ILLUSTRATIVE_NOT_FOR_FILING",FilingReady:false,
 TotalIncomeSen:r.DeclaredTotalIncomeSen,DraftReliefsExcludedSen:r.DraftReliefsSen,
 VerifiedReliefsAppliedSen:0,IllustrativeChargeableIncomeSen:r.DeclaredTotalIncomeSen,
 GrossTaxSen:tax.GrossTaxSen,EnteredRebatesSen:0,TaxAfterEnteredRebatesSen:tax.GrossTaxSen,
 PCBSen:r.PCBSen,IllustrativeBalanceSen:tax.GrossTaxSen-r.PCBSen,Brackets:tax.Lines,
 Warnings:[]string{
 "Total income is declared by the user; assessable and aggregate income are not validated.",
 "No relief is deducted; the draft amount is displayed only for reconciliation.",
 "All rebates are rejected pending verified statutory eligibility logic.",
 "Special dividend tax, tax credits, statutory rounding, joint assessments and other adjustments are not implemented.",
 "A negative indicative balance is not confirmation of an LHDN refund.",
 },
 },nil
}
