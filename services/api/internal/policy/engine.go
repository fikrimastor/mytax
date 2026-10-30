// Package policy implements deterministic arithmetic for draft relief rules.
// It does not interpret the law or certify tax entitlement.
package policy

import (
 "errors"
 "sort"
)

type Status string
const (
 Draft Status = "DRAFT"
 Approved Status = "APPROVED"
)
type Source struct {
 URL string `json:"url"`
 Section string `json:"section"`
 RetrievedOn string `json:"retrieved_on"`
}
type Rule struct {
 ID string `json:"id"`
 Group string `json:"group"`
 LimitSen int64 `json:"limit_sen"`
 RequiredFacts []string `json:"required_facts"`
 Sources []Source `json:"sources"`
}
type Group struct { ID string `json:"id"`; LimitSen int64 `json:"limit_sen"` }
type Package struct {
 AssessmentYear int `json:"assessment_year"`
 Version string `json:"version"`
 Status Status `json:"status"`
 Rules []Rule `json:"rules"`
 Groups []Group `json:"groups"`
}
type Claim struct {
 ID string `json:"id"`
 RuleID string `json:"rule_id"`
 AmountSen int64 `json:"amount_sen"`
 EvidencePresent bool `json:"evidence_present"`
 Duplicate bool `json:"duplicate"`
 Facts map[string]bool `json:"facts"`
}
type Decision struct {
 ClaimID string `json:"claim_id"`
 Status string `json:"status"`
 ProvisionalAmountSen int64 `json:"provisional_amount_sen"`
 ApprovedAmountSen int64 `json:"approved_amount_sen"`
 Reasons []string `json:"reasons"`
}
type Result struct {
 AssessmentYear int `json:"assessment_year"`
 PackageVersion string `json:"package_version"`
 PackageStatus Status `json:"package_status"`
 Decisions []Decision `json:"decisions"`
 ApprovedTotalSen int64 `json:"approved_total_sen"`
 Notice string `json:"notice"`
}
func Evaluate(p Package, claims []Claim)(Result,error){
 if p.AssessmentYear<2000 || p.AssessmentYear>2100 {return Result{},errors.New("invalid assessment year")}
 if p.Version=="" {return Result{},errors.New("package version required")}
 if p.Status!=Draft && p.Status!=Approved {return Result{},errors.New("unknown package status")}
 rules:=map[string]Rule{}; groups:=map[string]Group{}
 for _,g:=range p.Groups {if g.ID==""||g.LimitSen<0{return Result{},errors.New("invalid group")};if _,ok:=groups[g.ID];ok{return Result{},errors.New("duplicate group")};groups[g.ID]=g}
 for _,rule:=range p.Rules {
  if rule.ID==""||rule.LimitSen<0{return Result{},errors.New("invalid rule")}
  if _,ok:=rules[rule.ID];ok{return Result{},errors.New("duplicate rule")}
  if rule.Group!="" {if _,ok:=groups[rule.Group];!ok{return Result{},errors.New("unknown group")}}
  rules[rule.ID]=rule
 }
 // Stable allocation irrespective of client claim ordering; never double-count IDs.
 ordered:=append([]Claim(nil),claims...)
 sort.Slice(ordered,func(i,j int)bool{return ordered[i].ID<ordered[j].ID})
 usedRules:=map[string]int64{};usedGroups:=map[string]int64{}; seen:=map[string]bool{}
 res:=Result{AssessmentYear:p.AssessmentYear,PackageVersion:p.Version,PackageStatus:p.Status,Decisions:make([]Decision,0,len(ordered)),Notice:"Draft rules are not legally approved and never authorize a tax claim."}
 for _,c:=range ordered{
  if c.ID==""||c.AmountSen<=0{return Result{},errors.New("claim ID and positive amount required")}
  if seen[c.ID] {return Result{},errors.New("duplicate claim ID")};seen[c.ID]=true
  d:=Decision{ClaimID:c.ID,Status:"REVIEW_REQUIRED",Reasons:[]string{}}
  rule,ok:=rules[c.RuleID]
  if !ok {d.Reasons=append(d.Reasons,"UNKNOWN_RELIEF_RULE")}
  if c.Duplicate {d.Reasons=append(d.Reasons,"DUPLICATE_EXPENSE")}
  if !c.EvidencePresent {d.Reasons=append(d.Reasons,"EVIDENCE_MISSING")}
  if ok {
   for _,key:=range rule.RequiredFacts {if !c.Facts[key] {d.Reasons=append(d.Reasons,"UNVERIFIED_FACT:"+key)}}
  }
  if len(d.Reasons)==0 {
   available:=rule.LimitSen-usedRules[rule.ID]
   if rule.Group!="" {groupAvailable:=groups[rule.Group].LimitSen-usedGroups[rule.Group];if groupAvailable<available{available=groupAvailable}}
   if available<0 {available=0}
   allowed:=c.AmountSen;if allowed>available {allowed=available}
   d.ProvisionalAmountSen=allowed
   usedRules[rule.ID]+=allowed
   if rule.Group!="" {usedGroups[rule.Group]+=allowed}
   if p.Status==Approved {d.Status="POLICY_APPROVED_PRELIMINARY";d.ApprovedAmountSen=allowed;res.ApprovedTotalSen+=allowed}else{d.Reasons=append(d.Reasons,"DRAFT_POLICY_NOT_APPROVED")}
  }
  res.Decisions=append(res.Decisions,d)
 }
 return res,nil
}
