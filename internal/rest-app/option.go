package rest_app

import (
	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/provider/logging"
)

type RestAppParam struct {
	Config *app.Config
	Logger logging.Logger
	Server Server
}

type RestAppOption = func(*RestAppParam)

func WithConfig(c *app.Config) RestAppOption {
	return func(p *RestAppParam) {
		p.Config = c
	}
}

func WithLogger(l logging.Logger) RestAppOption {
	return func(p *RestAppParam) {
		p.Logger = l
	}
}

func WithServer(s Server) RestAppOption {
	return func(p *RestAppParam) {
		p.Server = s
	}
}
