package middleware

import (
	"github.com/labstack/echo"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
)

func WithScheduler(schedule *schedule.Scheduler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("schedule", schedule)
			return next(c)
		}
	}
}
func WithDataStore(store *db.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("store", store)
			return next(c)
		}
	}
}
