package assessment

import (
 "testing"
 "github.com/fikrimastor/mytax/services/api/internal/policy"
)
func TestDraftClaimsNeverReduceTax(t *testing.T){
 base,err:=Preview(Request{AssessmentYear:2025,DeclaredChargeableIncomeSen:7000000})
 if err!=nil{t.Fatal(err)}
 full,err:=Preview(Request{
  AssessmentYear:2025,DeclaredChargeableIncomeSen:7000000,
  Claims:[]policy.Claim{{ID:"internet",RuleID:"LIFESTYLE_INTERNET",AmountSen:12900,EvidencePresent:true,
   Facts:map[string]bool{"resident":true,"within_year":true,"internet_in_own_name":true,"qualifying_expense":true,"no_other_claim":true}}},
 })
 if err!=nil{t.Fatal(err)}
 if full.FilingReady || full.Policy.ApprovedTotalSen!=0 || full.GrossTax.GrossTaxSen!=base.GrossTax.GrossTaxSen {t.Fatalf("draft claim incorrectly affected tax: %+v",full)}
 if full.Reconciliation.DraftClaimsExcludedSen!=12900 || !full.Reconciliation.InvariantSatisfied{t.Fatalf("reconciliation failed: %+v",full)}
}
func TestInvalidYearsAndAmounts(t *testing.T){
 for _,r:=range []Request{{AssessmentYear:2026,DeclaredChargeableIncomeSen:7000000},{AssessmentYear:2025,DeclaredChargeableIncomeSen:-1}}{
  if _,err:=Preview(r);err==nil{t.Fatalf("accepted invalid request %+v",r)}
 }
}
func TestRejectedDuplicateClaimID(t *testing.T){
 c:=policy.Claim{ID:"same",RuleID:"LIFESTYLE_INTERNET",AmountSen:100}
 _,err:=Preview(Request{AssessmentYear:2025,Claims:[]policy.Claim{c,c}})
 if err==nil{t.Fatal("duplicate ID accepted")}
}
