package api

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	mid "github.com/labstack/echo/middleware"
	"github.com/sphiecoh/apimonitor/conf"
)

type Server struct {
	C *conf.Config
	H Handler
}

func (srv *Server) Start() {
	server := echo.New()
	server.Server.ReadTimeout = time.Second * 5
	server.Server.WriteTimeout = time.Second * 10
	server.Use(mid.Logger())
	server.Use(mid.Recover())
	server.POST("/", srv.H.CreateTest)
	server.GET("/", srv.H.GetAllTests)
	server.GET("/:id", srv.H.GetTestResult)
	server.DELETE("/:id", srv.H.DeleteTest)
	logrus.Fatal(server.Start(srv.C.Port))
}
