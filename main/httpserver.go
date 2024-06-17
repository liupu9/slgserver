package main

import (
	"log"

	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/server/httpserver/controller"
)

const httpServer string = "httpserver"

func main() {

	db.TestDB()

	e := echo.New()
	e.Use(mw.Recover())

	g := e.Group("")
	new(controller.AccountController).RegisterRoutes(g)
	e.Server.Addr = getHttpAddr()
	log.Fatal(e.StartServer(e.Server))
}

func getHttpAddr() string {
	host := config.File.MustValue(httpServer, "host", "")
	port := config.File.MustValue(httpServer, "port", "8088")
	return host + ":" + port
}
