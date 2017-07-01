package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/labstack/echo"
	mid "github.com/labstack/echo/middleware"
	"github.com/savaki/swag"
	"github.com/savaki/swag/endpoint"
	"github.com/savaki/swag/swagger"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/utahta/swagger-doc/assets"
)

type Server struct {
	C *conf.Config
	H Handler
}

//Start starts the webserver
func (srv *Server) Start() {
	server := echo.New()
	post := endpoint.New("post", "/tests", "Add a new test to the store",
		endpoint.Handler(srv.H.CreateTest),
		endpoint.Description("Additional information on adding a pet to the store"),
		endpoint.Body(db.ApiTest{}, "Test object that needs to be added", true),
		endpoint.Response(http.StatusCreated, db.ApiTest{}, "Successfully added test"),
	)
	getall := endpoint.New("get", "/tests", "Find all tests",
		endpoint.Handler(srv.H.GetAllTests),
		endpoint.Response(http.StatusOK, []db.ApiTest{}, "successful operation"),
	)
	get := endpoint.New("get", "/tests/{Id}/results", "Find all results by test ID",
		endpoint.Handler(srv.H.GetTestResult),
		endpoint.Path("Id", "string", "ID of the test", true),
		endpoint.Response(http.StatusOK, []db.ApiResult{}, "successful operation"),
	)
	del := endpoint.New("delete", "/{Id}", "Delete test by ID",
		endpoint.Handler(srv.H.DeleteTest),
		endpoint.Description("Delete test and its results"),
		endpoint.Path("Id", "string", "ID of the test", true),
		endpoint.Response(http.StatusOK, db.ApiTest{}, "Successfully delted test"),
	)
	api := swag.New(swag.Endpoints(post, getall, get, del))

	server.Server.ReadTimeout = time.Second * 5
	server.Server.WriteTimeout = time.Second * 10
	server.Use(mid.Logger())
	server.Use(mid.Recover())
	api.Walk(func(path string, endpoint *swagger.Endpoint) {
		h := endpoint.Handler.(func(c echo.Context) error)
		path = swag.ColonPath(path)

		switch strings.ToLower(endpoint.Method) {
		case "get":
			server.GET(path, h)
		case "head":
			server.HEAD(path, h)
		case "options":
			server.OPTIONS(path, h)
		case "delete":
			server.DELETE(path, h)
		case "put":
			server.PUT(path, h)
		case "post":
			server.POST(path, h)
		case "trace":
			server.TRACE(path, h)
		case "patch":
			server.PATCH(path, h)
		case "connect":
			server.CONNECT(path, h)
		}
	})
	enableCors := true
	fs := &assetfs.AssetFS{
		Asset:     assets.Asset,
		AssetDir:  assets.AssetDir,
		AssetInfo: assets.AssetInfo,
	}
	assetHandler := http.FileServer(fs)
	server.GET("/swagger", echo.WrapHandler(api.Handler(enableCors)))
	server.GET("/", echo.WrapHandler(assetHandler))
	server.GET("/css/*", echo.WrapHandler(assetHandler))
	server.GET("/lib/*", echo.WrapHandler(assetHandler))
	server.GET("/images/*", echo.WrapHandler(assetHandler))
	server.GET("/swagger-ui.js", echo.WrapHandler(assetHandler))
	logrus.Fatal(server.Start(srv.C.Port))
}
