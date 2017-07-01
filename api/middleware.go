package api

import (
	"net/http/httptest"

	"github.com/betacraft/yaag/middleware"
	"github.com/betacraft/yaag/yaag"
	"github.com/betacraft/yaag/yaag/models"
	"github.com/labstack/echo"
)

//Documentation middleware for api documentation
func Documentation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !yaag.IsOn() {
				return next(c)
			}
			req := c.Request()
			writer := httptest.NewRecorder()
			apiCall := models.ApiCall{}
			middleware.Before(&apiCall, req)
			next(c)
			middleware.After(&apiCall, writer, c.Response().Writer, req)
			return nil
		}
	}
}
