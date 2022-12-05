package queueing

import (
	"context"
)

type Pubsub interface {
	Publish(ctx context.Context, p PublishParam) error
	Subscribe(ctx context.Context, p SubscribeParam) error
}

type PublishParam struct {
	ExchangeName string
	MessageBody  []byte
}

type SubscribeParam struct {
	QueueName string
	Listener  Listener
}

type Listener = func(ctx context.Context, message Message) error

type Message interface {
	GetId() string
	GetBody() []byte
	Ack() error
	Nack() error
}
