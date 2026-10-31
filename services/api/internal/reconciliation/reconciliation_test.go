package reconciliation

import "testing"

func TestNoDraftReliefDeduction(t *testing.T){
 base,err:=Preview(Request{AssessmentYear:2025,DeclaredTotalIncomeSen:7000000,PCBSen:100000})
 if err!=nil{t.Fatal(err)}
 withDraft,err:=Preview(Request{AssessmentYear:2025,DeclaredTotalIncomeSen:7000000,DraftReliefsSen:600000,PCBSen:100000})
 if err!=nil{t.Fatal(err)}
 if base.GrossTaxSen!=370000 || withDraft.GrossTaxSen!=base.GrossTaxSen || withDraft.IllustrativeChargeableIncomeSen!=7000000 || withDraft.DraftReliefsExcludedSen!=600000{t.Fatalf("unsafe draft discount: %+v",withDraft)}
 if withDraft.IllustrativeBalanceSen!=270000 || withDraft.FilingReady{t.Fatal("incorrect preview result")}
}
func TestRefundDirectionIsNotOfficial(t *testing.T){
 r,err:=Preview(Request{AssessmentYear:2025,DeclaredTotalIncomeSen:7000000,PCBSen:500000})
 if err!=nil || r.IllustrativeBalanceSen!=-130000 || r.FilingReady {t.Fatalf("unexpected %+v %v",r,err)}
}
func TestRejectUnverifiedReliefsRebatesAndYears(t *testing.T){
 tests:=[]Request{
 {AssessmentYear:2026},{AssessmentYear:2025,VerifiedReliefsSen:1},
 {AssessmentYear:2025,EnteredRebatesSen:40000},
 {AssessmentYear:2025,PCBSen:-1},
 }
 for _,tt:=range tests {if _,err:=Preview(tt);err==nil{t.Fatalf("accepted %+v",tt)}}
}
