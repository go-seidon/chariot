package rabbitmq

import (
	"github.com/go-seidon/chariot/internal/rabbitmq"
)

type RabbitParam struct {
	Connection rabbitmq.Connection
}

type RabbitOption = func(*RabbitParam)

func WithConnection(conn rabbitmq.Connection) RabbitOption {
	return func(p *RabbitParam) {
		p.Connection = conn
	}
}
