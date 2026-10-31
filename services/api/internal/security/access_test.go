package security

import ("net/http/httptest";"testing")
func TestBearerFailsClosed(t *testing.T){
 for _,value:=range []string{"","Basic value","Bearer","Bearer a b","Bearer a,b"}{
  r:=httptest.NewRequest("GET","/",nil);if value!=""{r.Header.Set("Authorization",value)}
  if _,err:=Bearer(r);err==nil{t.Fatalf("accepted %q",value)}
 }
 r:=httptest.NewRequest("GET","/",nil);r.Header.Add("Authorization","Bearer one");r.Header.Add("Authorization","Bearer two")
 if _,err:=Bearer(r);err==nil{t.Fatal("multiple headers accepted")}
 r=httptest.NewRequest("GET","/",nil);r.Header.Set("Authorization","Bearer abc")
 s,err:=Bearer(r);if err!=nil||s!="abc"{t.Fatalf("valid header failed %s %v",s,err)}
}
