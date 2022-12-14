package queue

import (
	"context"

	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/chariot/internal/queuehandler"
	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/serialization"
)

type Queue interface {
	Start(ctx context.Context) error
}

type queue struct {
	queuer     queueing.Queuer
	serializer serialization.Serializer
	fileClient file.File
}

// @note for refactoring: use go-routine
func (q *queue) Start(ctx context.Context) error {
	var err error

	fileHandler := queuehandler.NewFile(queuehandler.FileParam{
		Serializer: q.serializer,
		File:       q.fileClient,
	})

	err = q.queuer.Open(ctx)
	if err != nil {
		return err
	}

	err = q.queuer.DeclareExchange(ctx, queueing.DeclareExchangeParam{
		ExchangeName: "file_replication",
		ExchangeType: queueing.EXCHANGE_FANOUT,
	})
	if err != nil {
		return err
	}

	que1, err := q.queuer.DeclareQueue(ctx, queueing.DeclareQueueParam{
		QueueName: "proceed_file_replication",
	})
	if err != nil {
		return err
	}

	err = q.queuer.BindQueue(ctx, queueing.BindQueueParam{
		ExchangeName: "file_replication",
		QueueName:    que1.Name,
	})
	if err != nil {
		return err
	}

	err = q.queuer.Subscribe(ctx, queueing.SubscribeParam{
		QueueName: que1.Name,
		Listener:  fileHandler.ProceedReplication,
	})
	if err != nil {
		return err
	}

	err = q.queuer.DeclareExchange(ctx, queueing.DeclareExchangeParam{
		ExchangeName: "file_deletion",
		ExchangeType: queueing.EXCHANGE_FANOUT,
	})
	if err != nil {
		return err
	}

	que2, err := q.queuer.DeclareQueue(ctx, queueing.DeclareQueueParam{
		QueueName: "proceed_file_deletion",
	})
	if err != nil {
		return err
	}

	err = q.queuer.BindQueue(ctx, queueing.BindQueueParam{
		ExchangeName: "file_deletion",
		QueueName:    que2.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

type QueueParam struct {
	Queuer     queueing.Queuer
	Serializer serialization.Serializer
	File       file.File
}

func NewQueue(p QueueParam) *queue {
	return &queue{
		queuer:     p.Queuer,
		serializer: p.Serializer,
		fileClient: p.File,
	}
}
