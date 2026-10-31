package evidence
import ("strings";"testing")
func TestDuplicatesAndAuditIntegrity(t *testing.T){
 a:=strings.Repeat("a",64);b:=strings.Repeat("b",64)
 items:=[]Item{{ID:"2",ClaimID:"c2",DocumentDigest:a,MerchantKey:"shop",InvoiceKey:"inv",AmountSen:5000},{ID:"1",ClaimID:"c1",DocumentDigest:a,MerchantKey:"shop",InvoiceKey:"inv",AmountSen:5000},{ID:"3",ClaimID:"c3",DocumentDigest:b,AmountSen:8000}}
 result,err:=Inspect(items);if err!=nil{t.Fatal(err)}
 if len(result.Findings)!=2||result.FilingReady||!VerifyAudit(items,result.Audit){t.Fatalf("unexpected result %+v",result)}
 result.Audit[1].SubjectID="tampered";if VerifyAudit(items,result.Audit){t.Fatal("tampering undetected")}
}
func TestInvalidEvidence(t *testing.T){
 ok:=Item{ID:"x",ClaimID:"c",DocumentDigest:strings.Repeat("a",64),AmountSen:100}
 for _,items:=range [][]Item{{ok,ok},{{ID:"bad",ClaimID:"c",DocumentDigest:"notasha",AmountSen:1}},{{ID:"bad",ClaimID:"c",DocumentDigest:strings.Repeat("a",64),AmountSen:-1}}}{
  if _,err:=Inspect(items);err==nil{t.Fatal("invalid input accepted")}
 }
}
