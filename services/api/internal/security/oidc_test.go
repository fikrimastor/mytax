package security
import ("context";"testing")
func TestOIDCConfigurationFailsClosed(t *testing.T){
 for _,tc:=range []struct{issuer,aud string}{{"http://issuer.test","mytax"},{"https://issuer.test",""},{"","x"}}{
  if _,err:=NewOIDCVerifier(context.Background(),tc.issuer,tc.aud);err==nil{t.Fatalf("accepted %q",tc.issuer)}
 }
 var verifier *Verifier
 if _,err:=verifier.Verify(context.Background(),"invalid");err==nil{t.Fatal("nil verifier accepted")}
}
