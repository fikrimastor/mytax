// Package eligibility screens candidate YA2025 relief claims. All results are
// provisional. No result authorizes a deduction or feeds filing calculations.
package eligibility

import ("errors";"math";"sort")

type Rule struct {ID string `json:"id"`; MaxSen int64 `json:"max_sen"`; Pool string `json:"pool"`; ExclusiveFamily string `json:"exclusive_family"`; RequiredFacts []string `json:"required_facts"`}
type Pool struct {ID string `json:"id"`; MaxSen int64 `json:"max_sen"`}
type Claim struct {ID string `json:"id"`; ExpenseKey string `json:"expense_key"`; RuleID string `json:"rule_id"`; AmountSen int64 `json:"amount_sen"`; EvidenceReference string `json:"evidence_reference"`; Facts map[string]bool `json:"facts"`}
type Decision struct {ID string `json:"id"`; State string `json:"state"`; CandidateSen int64 `json:"candidate_sen"`; Reasons []string `json:"reasons"`}
type Result struct {Year int `json:"assessment_year"`; Status string `json:"status"`; Decisions []Decision `json:"decisions"`; ProvisionalTotalSen int64 `json:"provisional_total_sen"`; ApprovedTotalSen int64 `json:"approved_total_sen"`; FilingReady bool `json:"filing_ready"`}
type Catalog struct {Year int; Rules []Rule; Pools []Pool}
func Draft2025() Catalog {return Catalog{Year:2025,Pools:[]Pool{{"LIFESTYLE",250000},{"EDUCATION",700000},{"MEDICAL",1000000}},Rules:[]Rule{
 {"LIFESTYLE_INTERNET",250000,"LIFESTYLE","",[]string{"resident","incurred_in_ya","own_name","qualifying_internet"}},
 {"LIFESTYLE_BOOKS",250000,"LIFESTYLE","",[]string{"resident","incurred_in_ya","qualifying_publication"}},
 {"EDUCATION_UPSKILL",200000,"EDUCATION","",[]string{"resident","incurred_in_ya","recognized_course"}},
 {"EDUCATION_OTHER",700000,"EDUCATION","",[]string{"resident","incurred_in_ya","qualifying_education"}},
 {"MEDICAL_VACCINATION",100000,"MEDICAL","",[]string{"resident","incurred_in_ya","qualifying_vaccine","qualifying_person"}},
 {"MEDICAL_DENTAL",100000,"MEDICAL","",[]string{"resident","incurred_in_ya","qualifying_dental","qualifying_person"}},
 {"FIRST_HOME_LOW",700000,"","FIRST_HOME",[]string{"resident","first_home","qualifying_spa_date","price_up_to_500k","qualifying_interest"}},
 {"FIRST_HOME_MID",500000,"","FIRST_HOME",[]string{"resident","first_home","qualifying_spa_date","price_above_500k_upto_750k","qualifying_interest"}},
 }}}
func Evaluate(cat Catalog, claims []Claim)(Result,error){
 if cat.Year!=2025{return Result{},errors.New("unsupported assessment year")}
 if len(claims)>1000{return Result{},errors.New("too many claims")}
 rules:=map[string]Rule{};pools:=map[string]int64{}
 for _,p:=range cat.Pools{if p.ID==""||p.MaxSen<0{return Result{},errors.New("invalid pool")};if _,ok:=pools[p.ID];ok{return Result{},errors.New("duplicate pool")};pools[p.ID]=p.MaxSen}
 for _,r:=range cat.Rules{if r.ID==""||r.MaxSen<0{return Result{},errors.New("invalid rule")};if _,ok:=rules[r.ID];ok{return Result{},errors.New("duplicate rule")};if r.Pool!=""{if _,ok:=pools[r.Pool];!ok{return Result{},errors.New("missing pool")}};rules[r.ID]=r}
 ordered:=append([]Claim(nil),claims...);sort.Slice(ordered,func(i,j int)bool{return ordered[i].ID<ordered[j].ID})
 seenID:=map[string]bool{};seenExpense:=map[string]bool{};family:=map[string]bool{};ruleUsed:=map[string]int64{};poolUsed:=map[string]int64{}
 out:=Result{Year:2025,Status:"DRAFT_NOT_FOR_FILING",Decisions:make([]Decision,0,len(claims)),FilingReady:false}
 for _,c:=range ordered{
  if c.ID==""||c.AmountSen<=0||seenID[c.ID]{return Result{},errors.New("invalid claim ID or amount")}
  seenID[c.ID]=true
  d:=Decision{ID:c.ID,State:"REVIEW_REQUIRED",Reasons:[]string{}}
  r,ok:=rules[c.RuleID];if !ok{d.Reasons=append(d.Reasons,"UNKNOWN_RULE")}
  if c.ExpenseKey==""{d.Reasons=append(d.Reasons,"EXPENSE_IDENTIFIER_MISSING")} else if seenExpense[c.ExpenseKey]{d.Reasons=append(d.Reasons,"DUPLICATE_EXPENSE")}
  if c.ExpenseKey!=""{seenExpense[c.ExpenseKey]=true}
  if c.EvidenceReference==""{d.Reasons=append(d.Reasons,"EVIDENCE_MISSING")}
  if ok {
   if r.ExclusiveFamily!="" && family[r.ExclusiveFamily]{d.Reasons=append(d.Reasons,"MUTUALLY_EXCLUSIVE_RELIEF")}
   for _,fact:=range r.RequiredFacts{if !c.Facts[fact]{d.Reasons=append(d.Reasons,"FACT_NOT_VERIFIED:"+fact)}}
  }
  if len(d.Reasons)==0 {
   remaining:=r.MaxSen-ruleUsed[r.ID]
   if r.Pool!=""{poolRemaining:=pools[r.Pool]-poolUsed[r.Pool];if poolRemaining<remaining{remaining=poolRemaining}}
   if remaining<0{remaining=0}
   allowed:=c.AmountSen;if allowed>remaining{allowed=remaining}
   d.CandidateSen=allowed;ruleUsed[r.ID]+=allowed;if r.Pool!=""{poolUsed[r.Pool]+=allowed}
   if allowed>0 && r.ExclusiveFamily!=""{family[r.ExclusiveFamily]=true}
   if out.ProvisionalTotalSen>math.MaxInt64-allowed{return Result{},errors.New("overflow")}
   out.ProvisionalTotalSen+=allowed
   d.Reasons=append(d.Reasons,"DRAFT_LEGAL_RULE_NOT_APPROVED")
   if allowed<c.AmountSen{d.Reasons=append(d.Reasons,"LIMIT_REACHED")}
  }
  out.Decisions=append(out.Decisions,d)
 }
 return out,nil
}
