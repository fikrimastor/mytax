package tax

import "errors"

type Claim struct {
 AssessmentYear int    `json:"assessment_year"`
 Category       string `json:"category"`
 AmountSen      int64  `json:"amount_sen"`
 IsDuplicate    bool   `json:"is_duplicate"`
 HasEvidence    bool   `json:"has_evidence"`
}
type Assessment struct {
 Status string `json:"status"`
 EligibleAmountSen int64 `json:"eligible_amount_sen"`
 Reasons []string `json:"reasons"`
 Disclaimer string `json:"disclaimer"`
}
// EvaluatePreliminary only checks general safety flags. It never approves a statutory claim.
func EvaluatePreliminary(c Claim) (Assessment,error) {
 if c.AssessmentYear<2000||c.AssessmentYear>2100 {return Assessment{},errors.New("invalid assessment year")}
 if c.Category=="" {return Assessment{},errors.New("category is required")}
 if c.AmountSen<=0 {return Assessment{},errors.New("amount_sen must be positive")}
 a:=Assessment{Status:"POLICY_NOT_APPROVED",EligibleAmountSen:0,Disclaimer:"Preliminary checks only. No legally approved assessment-year rule package is installed."}
 if c.IsDuplicate {a.Status="BLOCKED";a.Reasons=append(a.Reasons,"POSSIBLE_DUPLICATE_CLAIM")}
 if !c.HasEvidence {a.Reasons=append(a.Reasons,"EVIDENCE_REQUIRED")}
 a.Reasons=append(a.Reasons,"LHDN_RULE_REVIEW_REQUIRED")
 return a,nil
}
// ApplyCap is only arithmetic; it does not establish tax eligibility or a statutory cap.
func ApplyCap(amountSen,usedSen,capSen int64)(int64,error){
 if amountSen<0||usedSen<0||capSen<0{return 0,errors.New("negative values not allowed")}
 if usedSen>=capSen{return 0,nil}
 remaining:=capSen-usedSen
 if amountSen<remaining{return amountSen,nil}
 return remaining,nil
}
