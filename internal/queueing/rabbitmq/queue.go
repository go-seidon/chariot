package rabbitmq

import (
	"context"

	"github.com/go-seidon/chariot/internal/queueing"
	"github.com/go-seidon/chariot/internal/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitQueue struct {
	conn rabbitmq.Connection
}

func (que *rabbitQueue) DeclareQueue(ctx context.Context, p queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
	ch, err := que.conn.Channel()
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

	q, err := ch.QueueDeclare(p.QueueName, true, false, false, false, args)
	if err != nil {
		return nil, err
	}

	res := &queueing.DeclareQueueResult{
		Name: q.Name,
	}
	return res, nil
}

func NewQueueing(opts ...RabbitOption) *rabbitQueue {
	p := RabbitParam{}
	for _, opt := range opts {
		opt(&p)
	}

	return &rabbitQueue{
		conn: p.Connection,
	}
}
