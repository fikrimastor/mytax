package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// maxRequestBody bounds every request body. The largest legitimate request is a claim
// list at policy's 1,000-claim limit, well under 1 MB; anything larger gets HTTP 413
// before JSON decoding allocates it.
const maxRequestBody = "1M"

// useMiddleware installs the middleware shared by every route.
func useMiddleware(e *echo.Echo, corsRaw string) {
	e.Use(
		middleware.Recover(),
		middleware.BodyLimit(maxRequestBody),
		middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: corsOrigins(corsRaw)}),
	)
}
