package policy

import "testing"

func TestReleaseCannotActivateDraft(t *testing.T){
 p:=YA2025Draft()
 a:=ReleaseRecord{PackageVersion:p.Version,AssessmentYear:2025,ReviewerID:"reviewer",ReviewedAt:"2026-10-31",ApprovalReference:"example",SourceSnapshotSHA256:"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
 if err:=ValidateRelease(p,a);err==nil{t.Fatal("draft approved")}
 p.Status=Approved
 if err:=ValidateRelease(p,ReleaseRecord{});err==nil{t.Fatal("missing evidence accepted")}
 if _,err:=ActivePackage(2025);err==nil{t.Fatal("unregistered release activated")}
}
