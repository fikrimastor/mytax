package eligibility
import "testing"
func claim(id,key,rule string,amt int64) Claim {
 r:=Draft2025();f:=map[string]bool{};for _,v:=range r.Rules{if v.ID==rule{for _,x:=range v.RequiredFacts{f[x]=true}}}
 return Claim{ID:id,ExpenseKey:key,RuleID:rule,AmountSen:amt,EvidenceReference:"invoice",Facts:f}
}
func TestPoolCap(t *testing.T){
 r,e:=Evaluate(Draft2025(),[]Claim{claim("b","expense2","LIFESTYLE_INTERNET",200000),claim("a","expense1","LIFESTYLE_BOOKS",190000)})
 if e!=nil||r.ProvisionalTotalSen!=250000||r.ApprovedTotalSen!=0||r.FilingReady{t.Fatalf("%+v %v",r,e)}
}
func TestDoubleCountAndExclusivity(t *testing.T){
 a:=claim("a","same","LIFESTYLE_BOOKS",50000);b:=claim("b","same","LIFESTYLE_INTERNET",50000)
 r,e:=Evaluate(Draft2025(),[]Claim{a,b})
 if e!=nil||r.ProvisionalTotalSen!=50000||r.Decisions[1].CandidateSen!=0{t.Fatalf("%+v %v",r,e)}
 x:=claim("a","loan1","FIRST_HOME_LOW",300000);y:=claim("b","loan2","FIRST_HOME_MID",300000)
 r,e=Evaluate(Draft2025(),[]Claim{x,y})
 if e!=nil||r.ProvisionalTotalSen!=300000||r.Decisions[1].CandidateSen!=0{t.Fatalf("%+v %v",r,e)}
}
func TestMissingConditionsAndInvalid(t *testing.T){
 x:=claim("x","k","MEDICAL_VACCINATION",200000);delete(x.Facts,"qualifying_vaccine")
 r,e:=Evaluate(Draft2025(),[]Claim{x});if e!=nil||r.ProvisionalTotalSen!=0{t.Fatal(r,e)}
 x=claim("x","k","MEDICAL_VACCINATION",200000);r,e=Evaluate(Draft2025(),[]Claim{x})
 if e!=nil||r.ProvisionalTotalSen!=100000{t.Fatal(r,e)}
 if _,e:=Evaluate(Draft2025(),[]Claim{x,x});e==nil{t.Fatal("duplicate ID accepted")}
}
