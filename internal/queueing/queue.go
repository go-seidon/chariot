package queueing

import (
	"context"
)

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
