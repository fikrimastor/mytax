package policy

import "testing"

func TestRegistryDraftAndNoDuplicate(t *testing.T){
 entries:=ReliefRegistry2025()
 if len(entries)<25{t.Fatalf("registry too small: %d",len(entries))}
 seen:=map[string]bool{}; max:=map[string]int64{}
 for _,e:=range entries{
  if seen[e.ID]||e.MaxSen<=0||e.SourceURL!=OfficialReliefSource||e.Status!="DRAFT_REVIEW_REQUIRED"{t.Fatalf("invalid %+v",e)}
  seen[e.ID]=true;max[e.ID]=e.MaxSen
 }
 for _,e:=range entries{if e.ParentID!=""{if !seen[e.ParentID]||e.MaxSen>max[e.ParentID]{t.Fatalf("invalid nesting %+v",e)}}}
 if _,err:=ActivePackage(2025);err==nil{t.Fatal("draft unexpectedly activated")}
}
func TestKnownYA2025Caps(t *testing.T){
 caps:=map[string]int64{}
 for _,v:=range ReliefRegistry2025(){caps[v.ID]=v.MaxSen}
 for id,want:=range map[string]int64{"SELF":900000,"MEDICAL":1000000,"LIFESTYLE":250000,"MEDICAL_VACCINATION":100000,"SELF_UPSKILL":200000,"SSPN":800000,"SOCSO":35000,"CHILDCARE":300000,"DISABLED_SELF":700000}{
  if caps[id]!=want{t.Errorf("%s: got %d want %d",id,caps[id],want)}
 }
}
