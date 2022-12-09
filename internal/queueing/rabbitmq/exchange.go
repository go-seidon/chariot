package rabbitmq

import (
	"context"

	"github.com/go-seidon/chariot/internal/queueing"
)

func (que *rabbitQueue) DeclareExchange(ctx context.Context, p queueing.DeclareExchangeParam) error {
	ch, err := que.conn.Channel()
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

func (que *rabbitQueue) BindQueue(ctx context.Context, p queueing.BindQueueParam) error {
	ch, err := que.conn.Channel()
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
