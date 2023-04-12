package queue

import (
	"context"

	"github.com/go-seidon/chariot/internal/queuehandler"
	"github.com/go-seidon/chariot/internal/service"
	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/serialization"
)

type Queue interface {
	Start(ctx context.Context) error
}

type queue struct {
	queuer     queueing.Queuer
	serializer serialization.Serializer
	fileClient service.File
}

func (q *queue) Start(ctx context.Context) error {
	err := q.queuer.Open(ctx)
	if err != nil {
		return err
	}

	type exchange struct {
		Name string
		Type string
	}
	exchanges := []exchange{
		{
			Name: "file_replication",
			Type: queueing.EXCHANGE_FANOUT,
		},
		{
			Name: "file_deletion",
			Type: queueing.EXCHANGE_FANOUT,
		},
	}

	exchangeListener := make(chan error)
	for _, e := range exchanges {
		go func(e exchange) {
			err := q.queuer.DeclareExchange(ctx, queueing.DeclareExchangeParam{
				ExchangeName: e.Name,
				ExchangeType: e.Type,
			})
			exchangeListener <- err
		}(e)
	}

	type queue struct {
		Name string
	}
	queues := []queue{
		{
			Name: "proceed_file_replication",
		},
		{
			Name: "proceed_file_deletion",
		},
	}

	queueListener := make(chan error)
	for _, que := range queues {
		go func(que queue) {
			_, err := q.queuer.DeclareQueue(ctx, queueing.DeclareQueueParam{
				QueueName: que.Name,
			})
			queueListener <- err
		}(que)
	}

	for i := 0; i < len(exchanges); i++ {
		el := <-exchangeListener
		if el != nil {
			err = el
		}
	}
	for i := 0; i < len(queues); i++ {
		ql := <-queueListener
		if ql != nil {
			err = ql
		}
	}
	if err != nil {
		return err
	}

	type bindQueue struct {
		ExchangeName string
		QueueName    string
	}
	bindQueues := []bindQueue{
		{
			ExchangeName: "file_replication",
			QueueName:    "proceed_file_replication",
		},
		{
			ExchangeName: "file_deletion",
			QueueName:    "proceed_file_deletion",
		},
	}
	bindQueListener := make(chan error)
	for _, bq := range bindQueues {
		go func(bq bindQueue) {
			err := q.queuer.BindQueue(ctx, queueing.BindQueueParam{
				ExchangeName: bq.ExchangeName,
				QueueName:    bq.QueueName,
			})
			bindQueListener <- err
		}(bq)
	}

	for i := 0; i < len(bindQueues); i++ {
		bql := <-bindQueListener
		if bql != nil {
			err = bql
		}
	}
	if err != nil {
		return err
	}

	fileHandler := queuehandler.NewFile(queuehandler.FileParam{
		Serializer: q.serializer,
		File:       q.fileClient,
	})

	type subscriber struct {
		QueueName string
		Listener  queueing.Listener
	}
	subscribers := []subscriber{
		{
			QueueName: "proceed_file_replication",
			Listener:  fileHandler.ProceedReplication,
		},
		{
			QueueName: "proceed_file_deletion",
			Listener:  fileHandler.ProceedDeletion,
		},
	}
	subscriberListener := make(chan error)
	for _, sub := range subscribers {
		go func(sub subscriber) {
			err := q.queuer.Subscribe(ctx, queueing.SubscribeParam{
				QueueName: sub.QueueName,
				Listener:  sub.Listener,
			})
			subscriberListener <- err
		}(sub)
	}

	for i := 0; i < len(subscribers); i++ {
		sl := <-subscriberListener
		if sl != nil {
			err = sl
		}
	}
	if err != nil {
		return err
	}

	return nil
}

type QueueParam struct {
	Queuer     queueing.Queuer
	Serializer serialization.Serializer
	File       service.File
}

func NewQueue(p QueueParam) *queue {
	return &queue{
		queuer:     p.Queuer,
		serializer: p.Serializer,
		fileClient: p.File,
	}
}
