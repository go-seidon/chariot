package queueing

import (
	"context"
)

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
