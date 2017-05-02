package web

import (
	"github.com/kanosaki/dumper/common"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Config struct {
	Address string `yaml:"address"`
}

type Server struct {
	Echo *echo.Echo
	Conf Config
}

func New(conf *common.Config) (*Server, error) {
	var wc Config
	if err := conf.Unmarshal("web", &wc); err != nil {
		return nil, err
	}
	e := echo.New()
	e.Use(middleware.Recover())
	return &Server{
		Echo: e,
		Conf: wc,
	}, nil
}

func (w *Server) Start() error {
	return w.Echo.Start(w.Conf.Address)
}
