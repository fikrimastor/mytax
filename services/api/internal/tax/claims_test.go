package tax

import "testing"

func TestApplyCap(t *testing.T){
 cases:=[]struct{amount,used,cap,want int64}{{12900,0,250000,12900},{50000,240000,250000,10000},{50000,250000,250000,0},{50000,300000,250000,0}}
 for _,tc:=range cases {got,err:=ApplyCap(tc.amount,tc.used,tc.cap);if err!=nil||got!=tc.want{t.Fatalf("got %d err %v want %d",got,err,tc.want)}}
 if _,err:=ApplyCap(-1,0,100);err==nil{t.Fatal("expected error")}
}
func TestNeverAutoApprove(t *testing.T){
 a,err:=EvaluatePreliminary(Claim{AssessmentYear:2025,Category:"LIFESTYLE",AmountSen:12900,HasEvidence:true})
 if err!=nil||a.Status!="POLICY_NOT_APPROVED"||a.EligibleAmountSen!=0{t.Fatalf("unexpected %+v / %v",a,err)}
 b,err:=EvaluatePreliminary(Claim{AssessmentYear:2025,Category:"LIFESTYLE",AmountSen:12900,HasEvidence:true,IsDuplicate:true})
 if err!=nil||b.Status!="BLOCKED"||b.EligibleAmountSen!=0{t.Fatalf("unexpected %+v / %v",b,err)}
}
