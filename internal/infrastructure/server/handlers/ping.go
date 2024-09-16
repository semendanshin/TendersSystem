package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type PingHandler struct {
}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (p *PingHandler) Register(g *echo.Group) {
	g.GET("/ping", p.Ping)
}

func (p *PingHandler) Ping(e echo.Context) error {
	return e.String(http.StatusOK, "ok")
}
