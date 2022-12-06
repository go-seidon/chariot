package queueing

import (
	"context"
)

const (
	PROVIDER_RABBITMQ = "rabbitmq"
)

type Queueing interface {
	Manager
	Exchange
	Pubsub
	Queue
}

type Manager interface {
	Init(ctx context.Context) error
}

type Exchange interface {
	DeclareExchange(ctx context.Context, p DeclareExchangeParam) error
	BindQueue(ctx context.Context, p BindQueueParam) error
}

type DeclareExchangeParam struct {
	ExchangeName string
	ExchangeType string
}

type BindQueueParam struct {
	ExchangeName string
	QueueName    string
}

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

type Queue interface {
	DeclareQueue(ctx context.Context, p DeclareQueueParam) (*DeclareQueueResult, error)
}

type DeclareQueueParam struct {
	QueueName  string
	DeadLetter *DeclareQueueDeadLetter
}

type DeclareQueueResult struct {
	Name string
}

type DeclareQueueDeadLetter struct {
	ExchangeName string
	RoutingKey   string
}
