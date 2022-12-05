package rabbitmq

import (
	"context"

	"github.com/go-seidon/chariot/internal/queueing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitQueue struct {
	conn Connection
}

func (ex *rabbitQueue) DeclareQueue(ctx context.Context, p queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
	var ch Channel
	var err error

	ch, err = ex.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	args := amqp.Table{}
	if p.DeadLetter != nil {
		if p.DeadLetter.ExchangeName != "" {
			args["x-dead-letter-exchange"] = p.DeadLetter.ExchangeName
		}
		if p.DeadLetter.RoutingKey != "" {
			args["x-dead-letter-routing-key"] = p.DeadLetter.RoutingKey
		}
	}

	que, err := ch.QueueDeclare(p.QueueName, true, false, false, false, args)
	if err != nil {
		return nil, err
	}

	res := &queueing.DeclareQueueResult{
		Name: que.Name,
	}
	return res, nil
}

func NewQueue(opts ...RabbitOption) *rabbitQueue {
	p := RabbitParam{}
	for _, opt := range opts {
		opt(&p)
	}

	return &rabbitQueue{
		conn: p.Connection,
	}
}
