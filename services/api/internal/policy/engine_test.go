package policy

import "testing"

func facts(keys ...string)map[string]bool{m:=map[string]bool{};for _,k:=range keys{m[k]=true};return m}
func validClaim(id,rule string,amt int64)Claim{
 keys:=[]string{"resident","within_year","internet_in_own_name","qualifying_expense","no_other_claim","qualifying_reading_material","qualifying_beneficiary"}
 return Claim{ID:id,RuleID:rule,AmountSen:amt,EvidencePresent:true,Facts:facts(keys...)}
}
func TestDraftNeverApproves(t *testing.T){
 p:=YA2025Draft()
 r,err:=Evaluate(p,[]Claim{validClaim("a","LIFESTYLE_INTERNET",200000)})
 if err!=nil||r.ApprovedTotalSen!=0||r.Decisions[0].ApprovedAmountSen!=0||r.Decisions[0].ProvisionalAmountSen!=200000{t.Fatalf("draft leaked approval %+v %v",r,err)}
}
func TestSharedCapAndDeterminism(t *testing.T){
 p:=YA2025Draft();p.Status=Approved // arithmetic-only test fixture; does NOT signify tax-policy sign-off
 claims:=[]Claim{validClaim("b","LIFESTYLE_INTERNET",200000),validClaim("a","LIFESTYLE_BOOKS",180000)}
 a,err:=Evaluate(p,claims);if err!=nil{t.Fatal(err)}
 b,err:=Evaluate(p,[]Claim{claims[1],claims[0]});if err!=nil{t.Fatal(err)}
 if a.ApprovedTotalSen!=250000||b.ApprovedTotalSen!=250000||a.Decisions[0].ApprovedAmountSen!=180000||a.Decisions[1].ApprovedAmountSen!=70000{t.Fatalf("cap allocation wrong %+v %+v",a,b)}
}
func TestMissingEvidenceAndFacts(t *testing.T){
 p:=YA2025Draft()
 c:=validClaim("a","LIFESTYLE_INTERNET",50000);c.EvidencePresent=false
 r,_:=Evaluate(p,[]Claim{c});if r.Decisions[0].ProvisionalAmountSen!=0{t.Fatal("evidence missing must block")}
 c.EvidencePresent=true;delete(c.Facts,"internet_in_own_name")
 r,_=Evaluate(p,[]Claim{c});if r.Decisions[0].ProvisionalAmountSen!=0{t.Fatal("required fact missing must block")}
 c.Facts["internet_in_own_name"]=true;c.Duplicate=true
 r,_=Evaluate(p,[]Claim{c});if r.Decisions[0].ProvisionalAmountSen!=0{t.Fatal("duplicate must block")}
}
func TestRejectedInputs(t *testing.T){
 p:=YA2025Draft()
 if _,err:=Evaluate(p,[]Claim{{ID:"x",RuleID:"LIFESTYLE_INTERNET",AmountSen:-1}});err==nil{t.Fatal("negative amount accepted")}
 if _,err:=Evaluate(p,[]Claim{validClaim("x","LIFESTYLE_INTERNET",1),validClaim("x","LIFESTYLE_INTERNET",1)});err==nil{t.Fatal("duplicate ID accepted")}
}
