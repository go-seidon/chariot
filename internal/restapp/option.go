package restapp

import (
	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/queue"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/health"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/queueing"
)

type RestAppParam struct {
	Config       *app.Config
	Logger       logging.Logger
	Repository   repository.Provider
	Server       Server
	Queuer       queueing.Queuer
	Queue        queue.Queue
	HealthClient health.HealthCheck
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

func WithQueuer(queuer queueing.Queuer) RestAppOption {
	return func(p *RestAppParam) {
		p.Queuer = queuer
	}
}

func WithQueue(queue queue.Queue) RestAppOption {
	return func(p *RestAppParam) {
		p.Queue = queue
	}
}

func WithHealth(hc health.HealthCheck) RestAppOption {
	return func(p *RestAppParam) {
		p.HealthClient = hc
	}
}
