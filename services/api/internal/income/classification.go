// Package income provides an evidence-first YA2025 classification preflight.
// Classification is provisional and never determines statutory income.
package income

import (
 "errors"
 "math"
 "sort"
)
type Entry struct {
 ID string `json:"id"`
 Kind string `json:"kind"`
 AmountSen int64 `json:"amount_sen"`
 EvidenceReference string `json:"evidence_reference"`
}
type Decision struct {
 ID string `json:"id"`
 FormScope string `json:"form_scope"`
 Status string `json:"status"`
 Reasons []string `json:"reasons"`
}
type Result struct {
 AssessmentYear int `json:"assessment_year"`
 DeclaredTotalSen int64 `json:"declared_total_sen"`
 Decisions []Decision `json:"decisions"`
 FilingReady bool `json:"filing_ready"`
 Notice string `json:"notice"`
}
// These are intake categories, not official Form BE labels.
var supported=map[string]string{
 "EMPLOYMENT":"BE_CANDIDATE",
 "PENSION":"REVIEW",
 "DIVIDEND":"SPECIAL_REVIEW",
 "INTEREST":"REVIEW",
 "RENTAL":"REVIEW",
 "ROYALTY":"REVIEW",
 "OTHER":"REVIEW",
 "BUSINESS":"FORM_B_REVIEW",
}
func Classify(year int, entries []Entry)(Result,error){
 if year!=2025{return Result{},errors.New("only YA 2025 intake classification supported")}
 if len(entries)>1000{return Result{},errors.New("too many income entries")}
 ordered:=append([]Entry(nil),entries...)
 sort.Slice(ordered,func(i,j int)bool{return ordered[i].ID<ordered[j].ID})
 out:=Result{AssessmentYear:year,Decisions:make([]Decision,0,len(entries)),FilingReady:false,Notice:"Amounts and sources are unverified; do not treat this as statutory income or Form BE filing advice."}
 seen:=map[string]bool{}
 for _,e:=range ordered{
  if e.ID==""||seen[e.ID]||e.AmountSen<0{return Result{},errors.New("invalid or duplicated income entry")}
  seen[e.ID]=true
  if e.AmountSen>math.MaxInt64-out.DeclaredTotalSen{return Result{},errors.New("income total overflow")}
  out.DeclaredTotalSen+=e.AmountSen
  scope,ok:=supported[e.Kind]
  decision:=Decision{ID:e.ID,FormScope:scope,Status:"REVIEW_REQUIRED",Reasons:[]string{}}
  if !ok {decision.FormScope="UNKNOWN";decision.Reasons=append(decision.Reasons,"UNKNOWN_INCOME_CATEGORY")}
  if e.EvidenceReference==""{decision.Reasons=append(decision.Reasons,"EVIDENCE_REFERENCE_MISSING")}
  if e.Kind=="BUSINESS"{decision.Reasons=append(decision.Reasons,"FORM_B_SCOPE_REQUIRES_REVIEW")}
  if e.Kind=="DIVIDEND"{decision.Reasons=append(decision.Reasons,"DIVIDEND_SPECIAL_TREATMENT_REQUIRES_REVIEW")}
  if e.Kind=="EMPLOYMENT"{decision.Reasons=append(decision.Reasons,"EA_AND_STATUTORY_EMPLOYMENT_INCOME_RECONCILIATION_REQUIRED")}
  if len(decision.Reasons)==0{decision.Reasons=append(decision.Reasons,"STATUTORY_TREATMENT_NOT_VERIFIED")}
  out.Decisions=append(out.Decisions,decision)
 }
 return out,nil
}
