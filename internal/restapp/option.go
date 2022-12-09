package restapp

import (
	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/queueing"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/logging"
)

type RestAppParam struct {
	Config     *app.Config
	Logger     logging.Logger
	Repository repository.Provider
	Server     Server
	Queueing   queueing.Queueing
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

func WithRepository(r repository.Provider) RestAppOption {
	return func(p *RestAppParam) {
		p.Repository = r
	}
}

func WithServer(s Server) RestAppOption {
	return func(p *RestAppParam) {
		p.Server = s
	}
}

func WithQueueing(que queueing.Queueing) RestAppOption {
	return func(p *RestAppParam) {
		p.Queueing = que
	}
}
