package calculation

import "testing"

func TestOfficialYA2025Boundaries(t *testing.T){
 // Rates and cumulative amounts from HASiL official YA2025 table.
 cases:=[]struct{income,tax int64}{
 {0,0},{500000,0},{2000000,15000},{3500000,60000},
 {5000000,150000},{7000000,370000},{10000000,940000},
 {40000000,8440000},{60000000,13640000},
 {200000000,52840000},{200000100,52840300},
 }
 for _,tc:=range cases{
  r,err:=ResidentGrossTax(2025,tc.income)
  if err!=nil||r.GrossTaxSen!=tc.tax{t.Errorf("income=%d tax=%d want=%d err=%v",tc.income,r.GrossTaxSen,tc.tax,err)}
 }
}
func TestRejectUnsupportedYears(t *testing.T){
 for _,year:=range []int{2024,2026,2030}{
  if _,err:=ResidentGrossTax(year,1000000);err==nil{t.Errorf("unsupported year %d accepted",year)}
 }
}
func TestFractionalIncomeAndInvalid(t *testing.T){
 r,err:=ResidentGrossTax(2025,2000001)
 if err!=nil||r.GrossTaxSen!=15000{t.Fatalf("unexpected: %+v %v",r,err)}
 if _,err:=ResidentGrossTax(2025,-1);err==nil{t.Fatal("negative income accepted")}
}
func TestBracketProgression(t *testing.T){
 previous:=int64(-1)
 for income:=int64(0);income<=10000000;income+=10001 {
  r,err:=ResidentGrossTax(2025,income)
  if err!=nil||r.GrossTaxSen<previous{t.Fatalf("non-monotonic at %d",income)}
  previous=r.GrossTaxSen
 }
}
