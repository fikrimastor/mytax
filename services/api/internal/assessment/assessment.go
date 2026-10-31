// Package assessment joins draft claim screening with illustrative YA2025 gross tax.
// No result from this package is an official tax return or relief determination.
package assessment

import (
 "errors"
 "math"
 "github.com/fikrimastor/mytax/services/api/internal/calculation"
 "github.com/fikrimastor/mytax/services/api/internal/policy"
)
type Request struct {
 AssessmentYear int `json:"assessment_year"`
 DeclaredChargeableIncomeSen int64 `json:"declared_chargeable_income_sen"`
 Claims []policy.Claim `json:"claims"`
}
type Reconciliation struct {
 DeclaredChargeableIncomeSen int64 `json:"declared_chargeable_income_sen"`
 TaxCalculationInputSen int64 `json:"tax_calculation_input_sen"`
 DraftClaimsExcludedSen int64 `json:"draft_claims_excluded_sen"`
 InvariantSatisfied bool `json:"invariant_satisfied"`
}
type Response struct {
 Status string `json:"status"`
 Policy policy.Result `json:"policy"`
 GrossTax calculation.Result `json:"gross_tax"`
 Reconciliation Reconciliation `json:"reconciliation"`
 FilingReady bool `json:"filing_ready"`
 Warnings []string `json:"warnings"`
}
func Preview(req Request)(Response,error){
 if req.AssessmentYear!=2025{return Response{},errors.New("assessment preview only supports YA 2025")}
 if req.DeclaredChargeableIncomeSen<0 || req.DeclaredChargeableIncomeSen>math.MaxInt64/2{return Response{},errors.New("invalid chargeable income")}
 if len(req.Claims)>1000{return Response{},errors.New("too many claims")}
 screened,err:=policy.Evaluate(policy.YA2025Draft(),req.Claims)
 if err!=nil{return Response{},err}
 gross,err:=calculation.ResidentGrossTax(req.AssessmentYear,req.DeclaredChargeableIncomeSen)
 if err!=nil{return Response{},err}
 var excluded int64
 for _,d:=range screened.Decisions {
  if d.ProvisionalAmountSen>math.MaxInt64-excluded{return Response{},errors.New("claim amount overflow")}
  excluded+=d.ProvisionalAmountSen
 }
 rec:=Reconciliation{
  DeclaredChargeableIncomeSen:req.DeclaredChargeableIncomeSen,
  TaxCalculationInputSen:gross.ChargeableIncomeSen,
  DraftClaimsExcludedSen:excluded,
  InvariantSatisfied:gross.ChargeableIncomeSen==req.DeclaredChargeableIncomeSen && screened.ApprovedTotalSen==0,
 }
 if !rec.InvariantSatisfied{return Response{},errors.New("unsafe draft claim inclusion")}
 return Response{
  Status:"ILLUSTRATIVE_NOT_FOR_FILING",Policy:screened,GrossTax:gross,Reconciliation:rec,FilingReady:false,
  Warnings:[]string{
   "Chargeable income is user-declared and is not validated against Form BE.",
   "Draft claim amounts are excluded from the illustrative gross-tax calculation.",
   "Tax rebates, PCB, credits, statutory rounding and special treatments are not computed.",
   "This is not an official LHDN reconciliation or filing-ready assessment.",
  },
 },nil
}
