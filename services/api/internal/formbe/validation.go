// Package formbe implements source-linked field and preflight validation for YA2025.
// It does not certify a tax return or authorize tax rebates.
package formbe

import "errors"

type Field struct {
 Code string `json:"code"`
 Label string `json:"label"`
 SourcePage int `json:"source_page"`
 Stage string `json:"stage"`
}
const Source = "https://www.hasil.gov.my/media/pshpbomm/explanatorynotes_be2025_2.pdf"
var fields=[]Field{
 {"A6","Type of assessment",3,"PROFILE"},
 {"B1","Statutory employment income",3,"INCOME"},
 {"B2","Statutory dividend income",5,"INCOME"},
 {"B14","Chargeable income",8,"CALCULATION"},
 {"B15","Chargeable income subject to Part XXII",9,"SPECIAL_TAX"},
 {"B16","Tax on Part XXII income",9,"SPECIAL_TAX"},
 {"B17","Chargeable income subject to Part I",9,"CALCULATION"},
 {"B18","Income tax on Part I",9,"CALCULATION"},
 {"B19","Total income tax",9,"CALCULATION"},
 {"B21","Total rebate",9,"REBATE"},
}
func Fields(ya int)([]Field,error){
 if ya!=2025{return nil,errors.New("unsupported assessment year")}
 return append([]Field(nil),fields...),nil
}
type PreflightRequest struct {
 AssessmentYear int `json:"assessment_year"`
 ResidentConfirmed bool `json:"resident_confirmed"`
 NoBusinessIncomeConfirmed bool `json:"no_business_income_confirmed"`
 AssessmentType string `json:"assessment_type"`
 EmploymentIncomeSen int64 `json:"employment_income_sen"`
 DividendIncomeSen int64 `json:"dividend_income_sen"`
 AggregateIncomeSen int64 `json:"aggregate_income_sen"`
 ChargeableIncomeSen int64 `json:"chargeable_income_sen"`
 SpouseReliefVerified bool `json:"spouse_relief_verified"`
}
type Check struct { Code string `json:"code"`; State string `json:"state"`; Message string `json:"message"` }
type PreflightResult struct {
 AssessmentYear int `json:"assessment_year"`
 FilingReady bool `json:"filing_ready"`
 Checks []Check `json:"checks"`
 IndicativeSelfRebateSen int64 `json:"indicative_self_rebate_sen"`
 IndicativeSpouseRebateSen int64 `json:"indicative_spouse_rebate_sen"`
 RebateAppliedSen int64 `json:"rebate_applied_sen"`
 SourceURL string `json:"source_url"`
}
func Preflight(r PreflightRequest)(PreflightResult,error){
 if r.AssessmentYear!=2025{return PreflightResult{},errors.New("only YA 2025 supported")}
 if r.EmploymentIncomeSen<0||r.DividendIncomeSen<0||r.AggregateIncomeSen<0||r.ChargeableIncomeSen<0{return PreflightResult{},errors.New("negative monetary value")}
 if r.ChargeableIncomeSen>r.AggregateIncomeSen{return PreflightResult{},errors.New("chargeable income exceeds aggregate income")}
 res:=PreflightResult{AssessmentYear:2025,SourceURL:Source,Checks:[]Check{}}
 check:=func(code string,ok bool,msg string){state:="PASS";if !ok{state="REVIEW"};res.Checks=append(res.Checks,Check{code,state,msg})}
 check("RESIDENCY",r.ResidentConfirmed,"Confirm resident status for YA 2025 under section 7.")
 check("FORM_SCOPE",r.NoBusinessIncomeConfirmed,"Form BE is for residents who do not carry on business.")
 validType:=r.AssessmentType=="SEPARATE"||r.AssessmentType=="JOINT"||r.AssessmentType=="SPOUSE_NO_INCOME"||r.AssessmentType=="SINGLE"
 check("A6",validType,"Assessment type requires independent eligibility checks.")
 if r.DividendIncomeSen>0 {check("B15_B16",false,"Part XXII dividend apportionment and 2% calculation require review.")}
 if r.ChargeableIncomeSen<=3500000 && r.ResidentConfirmed && r.NoBusinessIncomeConfirmed && validType {
  res.IndicativeSelfRebateSen=40000
  if r.SpouseReliefVerified {res.IndicativeSpouseRebateSen=40000}
 }
 check("B21",false,"Indicative rebates are not applied: verified full statutory conditions and assessment context required.")
 check("FORM_BE_COMPLETE",false,"Full statutory income, relief, credit and payment crosswalk is incomplete.")
 return res,nil
}
