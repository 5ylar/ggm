package main

import (
	"log"
	"net/http/httputil"

	"github.com/labstack/echo/v4"
)

func ProxyRequest(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer

	log.Println("r.URL", r.URL)

	proxy := httputil.NewSingleHostReverseProxy(c.Request().URL)

	proxy.ServeHTTP(w, r)

	return nil
}

func main() {
	app := echo.New()
	app.Any("*", ProxyRequest)
	app.Start("0.0.0.0:8080")
}
