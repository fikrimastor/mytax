package income

import "testing"

func TestIncomeIntakeNeverCertifies(t *testing.T){
 r,err:=Classify(2025,[]Entry{{ID:"b",Kind:"BUSINESS",AmountSen:50000,EvidenceReference:"statement"},{ID:"a",Kind:"EMPLOYMENT",AmountSen:7000000,EvidenceReference:"EA"}})
 if err!=nil||r.FilingReady||r.DeclaredTotalSen!=7050000||r.Decisions[0].ID!="a"||r.Decisions[1].FormScope!="FORM_B_REVIEW"{t.Fatalf("unexpected result %+v, %v",r,err)}
}
func TestUnknownAndDividendReview(t *testing.T){
 r,err:=Classify(2025,[]Entry{{ID:"x",Kind:"DIVIDEND",AmountSen:1000},{ID:"y",Kind:"CRYPTO",AmountSen:2000}})
 if err!=nil||r.Decisions[0].Status!="REVIEW_REQUIRED"||r.Decisions[1].FormScope!="UNKNOWN"{t.Fatalf("unexpected result %+v %v",r,err)}
}
func TestInvalidIncome(t *testing.T){
 cases:=[][]Entry{{{ID:"x",AmountSen:-1}},{{ID:"x"},{ID:"x"}},{{ID:"x",AmountSen:9223372036854775807},{ID:"y",AmountSen:1}}}
 for _,c:=range cases{if _,err:=Classify(2025,c);err==nil{t.Fatalf("accepted invalid %+v",c)}}
 if _,err:=Classify(2026,nil);err==nil{t.Fatal("unsupported YA accepted")}
}
