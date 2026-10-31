// Package calculation computes indicative resident individual tax on an already
// established chargeable income. It does not determine chargeable income.
package calculation

import (
 "errors"
 "math"
)

type Band struct {
 UpToSen int64 `json:"up_to_sen"`
 RatePercent int64 `json:"rate_percent"`
}
type Line struct {
 FromSen int64 `json:"from_sen"`
 ToSen int64 `json:"to_sen"`
 RatePercent int64 `json:"rate_percent"`
 TaxSen int64 `json:"tax_sen"`
}
type Result struct {
 AssessmentYear int `json:"assessment_year"`
 ChargeableIncomeSen int64 `json:"chargeable_income_sen"`
 GrossTaxSen int64 `json:"gross_tax_sen"`
 Lines []Line `json:"lines"`
 Status string `json:"status"`
 SourceURL string `json:"source_url"`
 Notice string `json:"notice"`
}
const SourceURL = "https://www.hasil.gov.my/individu/kadar-cukai/"
// 0 upper bound is a sentinel for the unbounded final band.
// These marginal rates are from HASiL resident-individual YA 2025 table.
var ya2025 = []Band{
 {500000,0},{2000000,1},{3500000,3},{5000000,6},
 {7000000,11},{10000000,19},{40000000,25},
 {60000000,26},{200000000,28},{0,30},
}
func Rates(ya int)([]Band,error){
 if ya!=2025{return nil,errors.New("tax rates not verified for requested assessment year")}
 return append([]Band(nil),ya2025...),nil
}
// ResidentGrossTax expects a caller-supplied, already established CHARGEABLE
// income. No relief deductions, rebates, special taxes, rounding to whole ringgit,
// PCB credits or filing liability are computed by this illustrative endpoint.
// Rate computation rounds each marginal band to the nearest sen (half up).
func ResidentGrossTax(ya int,chargeableIncomeSen int64)(Result,error){
 if chargeableIncomeSen<0{return Result{},errors.New("income must not be negative")}
 bands,err:=Rates(ya);if err!=nil{return Result{},err}
 out:=Result{AssessmentYear:ya,ChargeableIncomeSen:chargeableIncomeSen,Status:"ILLUSTRATIVE_GROSS_TAX_ONLY",SourceURL:SourceURL,Notice:"Not a filing-ready tax calculation. Chargeable income, statutory rounding, rebates, credits, PCB and applicable special cases are not validated."}
 lower:=int64(0)
 for _,b:=range bands{
  if chargeableIncomeSen<=lower {break}
  upper:=b.UpToSen;if upper==0||upper>chargeableIncomeSen{upper=chargeableIncomeSen}
  delta:=upper-lower
  if delta<=0{continue}
  if b.RatePercent>0 && delta>(math.MaxInt64-50)/b.RatePercent {return Result{},errors.New("income too large")}
  tax:=(delta*b.RatePercent+50)/100
  if out.GrossTaxSen>math.MaxInt64-tax{return Result{},errors.New("tax overflow")}
  out.Lines=append(out.Lines,Line{FromSen:lower,ToSen:upper,RatePercent:b.RatePercent,TaxSen:tax})
  out.GrossTaxSen+=tax
  lower=upper
 }
 return out,nil
}
