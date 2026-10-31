package formbe

import "testing"

func TestFormBEPreflightThreshold(t *testing.T){
 r:=PreflightRequest{AssessmentYear:2025,ResidentConfirmed:true,NoBusinessIncomeConfirmed:true,AssessmentType:"SINGLE",AggregateIncomeSen:3500000,ChargeableIncomeSen:3500000}
 got,err:=Preflight(r);if err!=nil{t.Fatal(err)}
 if got.IndicativeSelfRebateSen!=40000||got.RebateAppliedSen!=0||got.FilingReady{t.Fatalf("bad rebate %+v",got)}
 r.ChargeableIncomeSen=3500001;r.AggregateIncomeSen=3500001
 got,err=Preflight(r);if err!=nil||got.IndicativeSelfRebateSen!=0{t.Fatalf("threshold incorrect %+v %v",got,err)}
}
func TestRebateNeedsConfirmation(t *testing.T){
 r:=PreflightRequest{AssessmentYear:2025,AssessmentType:"SINGLE",AggregateIncomeSen:2500000,ChargeableIncomeSen:2500000}
 got,err:=Preflight(r);if err!=nil||got.IndicativeSelfRebateSen!=0||got.RebateAppliedSen!=0{t.Fatalf("unsafe rebate %+v",got)}
}
func TestDividendsFlagged(t *testing.T){
 r:=PreflightRequest{AssessmentYear:2025,ResidentConfirmed:true,NoBusinessIncomeConfirmed:true,AssessmentType:"SINGLE",AggregateIncomeSen:7000000,ChargeableIncomeSen:5000000,DividendIncomeSen:100000}
 got,err:=Preflight(r);if err!=nil{t.Fatal(err)}
 found:=false;for _,c:=range got.Checks{if c.Code=="B15_B16" && c.State=="REVIEW"{found=true}}
 if !found||got.FilingReady{t.Fatalf("dividend omission %+v",got)}
}
func TestRejectInvalid(t *testing.T){
 for _,r:=range []PreflightRequest{{AssessmentYear:2026},{AssessmentYear:2025,AggregateIncomeSen:100,ChargeableIncomeSen:101},{AssessmentYear:2025,EmploymentIncomeSen:-1}}{
  if _,err:=Preflight(r);err==nil{t.Fatalf("accepted %+v",r)}
 }
}
