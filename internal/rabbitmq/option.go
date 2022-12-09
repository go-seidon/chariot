package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqParam struct {
	Address    string
	Connection *amqp.Connection
}

type RabbitmqOption = func(*RabbitmqParam)

func WithAddress(addr string) RabbitmqOption {
	return func(p *RabbitmqParam) {
		p.Address = addr
	}
}

func WithConnection(conn *amqp.Connection) RabbitmqOption {
	return func(p *RabbitmqParam) {
		p.Connection = conn
	}
}
