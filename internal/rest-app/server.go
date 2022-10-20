package rest_app

import (
	"context"

	"github.com/labstack/echo/v4"
)

type Server interface {
	Start(address string) error
	Shutdown(ctx context.Context) error
}

type echoServer struct {
	e *echo.Echo
}

func (s *echoServer) Start(address string) error {
	return s.e.Start(address)
}

func (s *echoServer) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
