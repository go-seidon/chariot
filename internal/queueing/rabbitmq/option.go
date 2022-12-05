package rabbitmq

type RabbitParam struct {
	Connection Connection
}

type RabbitOption = func(*RabbitParam)

func WithConnection(conn Connection) RabbitOption {
	return func(p *RabbitParam) {
		p.Connection = conn
	}
}
