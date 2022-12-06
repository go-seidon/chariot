package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection interface {
	// @note: call init on app startup
	Init(ctx context.Context) error

	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
	NotifyBlocked(receiver chan amqp.Blocking) chan amqp.Blocking
	Close() error
	IsClosed() bool
	Channel() (*amqp.Channel, error)
}

type Channel interface {
	Close() error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
}

type connection struct {
	conn *amqp.Connection
	addr string
}

func (c *connection) Init(ctx context.Context) error {
	conn, err := amqp.Dial(c.addr)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *connection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return c.conn.NotifyClose(receiver)
}

func (c *connection) NotifyBlocked(receiver chan amqp.Blocking) chan amqp.Blocking {
	return c.conn.NotifyBlocked(receiver)
}

func (c *connection) Close() error {
	return c.conn.Close()
}

func (c *connection) IsClosed() bool {
	return c.conn.IsClosed()
}

func (c *connection) Channel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

func NewConnection(opts ...RabbitmqOption) *connection {
	p := RabbitmqParam{}
	for _, opt := range opts {
		opt(&p)
	}

	conn := p.Connection
	if conn == nil {
		conn = &amqp.Connection{}
	}

	return &connection{
		addr: p.Address,
		conn: conn,
	}
}
