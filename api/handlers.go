// @SubApi Test API [/]
package api

import (
	"net/http"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
	"gopkg.in/robfig/cron.v2"
)

type Handler struct {
	S     *schedule.Scheduler
	Store *db.Store
}

func (h *Handler) CreateTest(c echo.Context) error {
	newtest := &db.ApiTest{}
	newtest.ID = db.GenerateID()
	c.Bind(newtest)
	data, err := json.Marshal(newtest)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if e := h.Store.Put(newtest.Name, h.Store.TestBucket, data); e != nil {
		logrus.Error(e)
		return e
	}
	job := schedule.ToJob(newtest, h.Store, h.S.Config)

	schedule, err := cron.Parse(newtest.Cron)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Test has invalid cron %s (%s)", newtest.Cron, newtest.Name))
	}
	h.S.Cron.Schedule(schedule, job)
	logrus.Infof("New test created %s", newtest.Name)
	return c.JSON(http.StatusCreated, nil)
}

// @Title GetAllTests
// @Description get all tests
// @Accept  json
// @Produce  json
// @Success 200 {object} []db.ApiTest
// @Router / [get]
func (h *Handler) GetAllTests(c echo.Context) error {

	result, err := h.Store.GetAllTests()
	if err != nil {
		return err
	}
	c.JSON(200, result)
	return nil
}
func (h *Handler) GetTestResult(c echo.Context) error {
	id := c.Param("id")
	result, err := h.Store.GetResultsByTest(id)
	if err != nil {
		return err
	}
	c.JSON(200, result)
	return nil
}
func (h *Handler) DeleteTest(c echo.Context) error {
	id := c.Param("id")

	if err := h.Store.DeleteTest(id); err != nil {
		return err
	}
	eid := h.S.Entries[id]
	h.S.Cron.Remove(eid)
	c.JSON(200, nil)
	return nil

}
func Index(c echo.Context) error {

	c.Redirect(301, "/web/html/layout.html")
	return nil

}
