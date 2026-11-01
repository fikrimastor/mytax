package main

import (
 "net/http"
 "context"
 "time"
 "os"
 "strconv"
 "github.com/labstack/echo/v4"
 "github.com/labstack/echo/v4/middleware"
 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/fikrimastor/mytax/services/api/internal/tax"
 "github.com/fikrimastor/mytax/services/api/internal/policy"
 "github.com/fikrimastor/mytax/services/api/internal/calculation"
 "github.com/fikrimastor/mytax/services/api/internal/assessment"
 "github.com/fikrimastor/mytax/services/api/internal/reconciliation"
 "github.com/fikrimastor/mytax/services/api/internal/formbe"
 "github.com/fikrimastor/mytax/services/api/internal/income"
 "github.com/fikrimastor/mytax/services/api/internal/eligibility"
 "github.com/fikrimastor/mytax/services/api/internal/evidence"
 "github.com/fikrimastor/mytax/services/api/internal/security"
 "github.com/fikrimastor/mytax/services/api/internal/persistence"
)
func main(){
 e:=echo.New()
 e.HideBanner=true
 e.Use(middleware.Recover(),middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins:corsOrigins(os.Getenv("MYTAX_CORS_ORIGINS"))}))
 if os.Getenv("MYTAX_ENABLE_AUTH_PREVIEW")=="true" {
  issuer:=os.Getenv("MYTAX_OIDC_ISSUER")
  audience:=os.Getenv("MYTAX_OIDC_AUDIENCE")
  databaseURL:=os.Getenv("DATABASE_URL")
  allowedClient:=os.Getenv("MYTAX_OIDC_ALLOWED_CLIENT")
  requiredScope:=os.Getenv("MYTAX_OIDC_REQUIRED_SCOPE")
  if issuer==""||audience==""||databaseURL==""||allowedClient==""||requiredScope=="" {e.Logger.Fatal("auth preview requires issuer, audience, and DATABASE_URL")}
  // Config errors still crash at boot; only the network discovery is lazy so
  // the API can start before Keycloak is ready. Until it succeeds, 503.
  if err:=security.ValidateAccessConfig(issuer,audience,allowedClient,requiredScope);err!=nil {e.Logger.Fatal(err)}
  verifier,err:=security.NewLazyVerifier(func(ctx context.Context)(security.TokenVerifier,error){
   v,err:=security.NewAccessVerifier(ctx,issuer,audience,allowedClient,requiredScope)
   if err!=nil{return nil,err}
   return v,nil
  },10*time.Second,5*time.Second)
  if err!=nil {e.Logger.Fatal(err)}
  pool,err:=pgxpool.New(context.Background(),databaseURL)
  if err!=nil {e.Logger.Fatal(err)}
  defer pool.Close()
  tenantDB,err:=persistence.NewTenantDB(pool,issuer,audience)
  if err!=nil {e.Logger.Fatal(err)}
  e.GET("/v1/secure/tenants/:tenant_id/evidence/count",security.RequireAccess(verifier,func(c echo.Context)error{
   id,ok:=c.Get("verified_identity").(security.Identity)
   if !ok{return c.JSON(http.StatusUnauthorized,map[string]string{"error":"unauthorized"})}
   count,err:=tenantDB.CountEvidence(c.Request().Context(),id,c.Param("tenant_id"))
   if err!=nil{return c.JSON(http.StatusForbidden,map[string]string{"error":"access denied"})}
   return c.JSON(http.StatusOK,map[string]any{"count":count,"status":"PROTOTYPE_METADATA_ONLY"})
  }))
 }
 e.GET("/healthz",func(c echo.Context)error{return c.JSON(http.StatusOK,map[string]string{"status":"ok"})})
 e.GET("/v1/policy/status",func(c echo.Context)error{
  ya,err:=strconv.Atoi(c.QueryParam("ya"))
  if err!=nil||ya<2000||ya>2100{return c.JSON(http.StatusBadRequest,map[string]string{"error":"valid ya required"})}
  return c.JSON(http.StatusOK,map[string]any{"assessment_year":ya,"status":"NOT_APPROVED","tax_calculation_enabled":false,"reason":"Official tax-year rules have not been signed off"})
 })
 e.POST("/v1/evidence/inspect",func(c echo.Context)error{
  if os.Getenv("MYTAX_ENABLE_EVIDENCE_PREVIEW")!="true" {return c.JSON(http.StatusNotFound,map[string]string{"error":"prototype disabled"})}
  var input struct{Items []evidence.Item `json:"items"`}
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  output,err:=evidence.Inspect(input.Items)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,output)
 })
 e.POST("/v1/eligibility/preview",func(c echo.Context)error{
  var input struct {AssessmentYear int `json:"assessment_year"`; Claims []eligibility.Claim `json:"claims"`}
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  if input.AssessmentYear!=2025{return c.JSON(http.StatusBadRequest,map[string]string{"error":"only YA 2025 supported"})}
  result,err:=eligibility.Evaluate(eligibility.Draft2025(),input.Claims)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
 e.POST("/v1/income/classify",func(c echo.Context)error{
  var input struct {AssessmentYear int `json:"assessment_year"`; Entries []income.Entry `json:"entries"`}
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  result,err:=income.Classify(input.AssessmentYear,input.Entries)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
 e.GET("/v1/formbe/fields",func(c echo.Context)error{
  ya,err:=strconv.Atoi(c.QueryParam("ya"))
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid year"})}
  result,err:=formbe.Fields(ya)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,map[string]any{"assessment_year":ya,"source_url":formbe.Source,"status":"PARTIAL_NOT_FOR_FILING","fields":result})
 })
 e.POST("/v1/formbe/preflight",func(c echo.Context)error{
  var input formbe.PreflightRequest
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  result,err:=formbe.Preflight(input)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
 e.POST("/v1/reconciliation/preview",func(c echo.Context)error{
  var input reconciliation.Request
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  result,err:=reconciliation.Preview(input)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
 e.POST("/v1/assessment/preview",func(c echo.Context)error{
  var input assessment.Request
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  out,err:=assessment.Preview(input)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,out)
 })
 e.GET("/v1/calculation/resident-gross",func(c echo.Context)error{
  ya,yearErr:=strconv.Atoi(c.QueryParam("ya"))
  income,incomeErr:=strconv.ParseInt(c.QueryParam("chargeable_income_sen"),10,64)
  if yearErr!=nil||incomeErr!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"valid ya and chargeable_income_sen required"})}
  result,err:=calculation.ResidentGrossTax(ya,income)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
 })
 e.GET("/v1/policy/registry",func(c echo.Context)error{
  if c.QueryParam("ya")!="2025" {return c.JSON(http.StatusNotFound,map[string]string{"error":"registry not available for year"})}
  return c.JSON(http.StatusOK,map[string]any{"assessment_year":2025,"status":"DRAFT_REVIEW_REQUIRED","entries":policy.ReliefRegistry2025(),"filing_ready":false})
 })
 e.GET("/v1/policy/catalog",func(c echo.Context)error{
  if c.QueryParam("ya")!="2025" {return c.JSON(http.StatusNotFound,map[string]string{"error":"no catalog for assessment year"})}
  return c.JSON(http.StatusOK,policy.YA2025Draft())
 })
 e.POST("/v1/claims/evaluate",func(c echo.Context)error{
  var input tax.Claim
  if err:=c.Bind(&input);err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"invalid JSON"})}
  a,err:=tax.EvaluatePreliminary(input)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,a)
 })
 port:=os.Getenv("PORT");if port==""{port="8080"}
 e.Logger.Fatal(e.Start(":"+port))
}
