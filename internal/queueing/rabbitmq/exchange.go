package rabbitmq

import (
	"context"

	"github.com/go-seidon/chariot/internal/queueing"
)

const (
	EXCHANGE_DIRECT  = "direct"
	EXCHANGE_FANOUT  = "fanout" //broadcasts all the messages it receives to all the queues it knows
	EXCHANGE_TOPIC   = "topic"
	EXCHANGE_HEADERS = "headers"
)

type rabbitExchange struct {
	conn Connection
}

func (ex *rabbitExchange) DeclareExchange(ctx context.Context, p queueing.DeclareExchangeParam) error {
	var ch Channel
	var err error

	ch, err = ex.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(p.ExchangeName, p.ExchangeType, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func (ex *rabbitExchange) BindQueue(ctx context.Context, p queueing.BindQueueParam) error {
	var ch Channel
	var err error

	ch, err = ex.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.QueueBind(p.QueueName, "", p.ExchangeName, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func NewExchange(opts ...RabbitOption) *rabbitExchange {
	p := RabbitParam{}
	for _, opt := range opts {
		opt(&p)
	}

	return &rabbitExchange{
		conn: p.Connection,
	}
}
