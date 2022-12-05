package rabbitmq

import (
	"context"

	"github.com/go-seidon/chariot/internal/queueing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitPubsub struct {
	conn Connection
}

func (pub *rabbitPubsub) Publish(ctx context.Context, p queueing.PublishParam) error {
	var ch Channel
	var err error

	ch, err = pub.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.PublishWithContext(ctx, p.ExchangeName, "", false, false, amqp.Publishing{
		Body:         p.MessageBody,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		return err
	}

	return nil
}

func (pub *rabbitPubsub) Subscribe(ctx context.Context, p queueing.SubscribeParam) error {
	var ch Channel
	var err error

	ch, err = pub.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	delivery, err := ch.Consume(p.QueueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	var forever chan struct{}
	go func() {
		for d := range delivery {
			p.Listener(ctx, &message{d: d})
		}
	}()
	<-forever

	return nil
}

func NewPubsub(opts ...RabbitOption) *rabbitPubsub {
	p := RabbitParam{}
	for _, opt := range opts {
		opt(&p)
	}

	return &rabbitPubsub{
		conn: p.Connection,
	}
}
