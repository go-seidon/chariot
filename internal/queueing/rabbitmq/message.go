package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type message struct {
	d amqp.Delivery
}

func (m *message) GetId() string {
	return m.d.MessageId
}

func (m *message) GetBody() []byte {
	return m.d.Body
}

func (m *message) Ack() error {
	return m.d.Ack(true)
}

func (m *message) Nack() error {
	return m.d.Nack(true, true)
}
