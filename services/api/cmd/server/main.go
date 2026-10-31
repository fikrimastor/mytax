package main

import (
 "net/http"
 "os"
 "strconv"
 "github.com/labstack/echo/v4"
 "github.com/labstack/echo/v4/middleware"
 "github.com/fikrimastor/mytax/services/api/internal/tax"
 "github.com/fikrimastor/mytax/services/api/internal/policy"
 "github.com/fikrimastor/mytax/services/api/internal/calculation"
)
func main(){
 e:=echo.New()
 e.HideBanner=true
 e.Use(middleware.Recover(),middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins:[]string{"http://localhost:3000"}}))
 e.GET("/healthz",func(c echo.Context)error{return c.JSON(http.StatusOK,map[string]string{"status":"ok"})})
 e.GET("/v1/policy/status",func(c echo.Context)error{
  ya,err:=strconv.Atoi(c.QueryParam("ya"))
  if err!=nil||ya<2000||ya>2100{return c.JSON(http.StatusBadRequest,map[string]string{"error":"valid ya required"})}
  return c.JSON(http.StatusOK,map[string]any{"assessment_year":ya,"status":"NOT_APPROVED","tax_calculation_enabled":false,"reason":"Official tax-year rules have not been signed off"})
 })
 e.GET("/v1/calculation/resident-gross",func(c echo.Context)error{
  ya,yearErr:=strconv.Atoi(c.QueryParam("ya"))
  income,incomeErr:=strconv.ParseInt(c.QueryParam("chargeable_income_sen"),10,64)
  if yearErr!=nil||incomeErr!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":"valid ya and chargeable_income_sen required"})}
  result,err:=calculation.ResidentGrossTax(ya,income)
  if err!=nil{return c.JSON(http.StatusBadRequest,map[string]string{"error":err.Error()})}
  return c.JSON(http.StatusOK,result)
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
