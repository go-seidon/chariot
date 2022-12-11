package app

import (
	"fmt"

	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/queueing/rabbitmq"
	conn "github.com/go-seidon/provider/rabbitmq"
)

func NewDefaultQueueing(config *Config) (queueing.Queueing, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if config.QueueProvider != queueing.PROVIDER_RABBITMQ {
		return nil, fmt.Errorf("invalid queue provider")
	}

	var que queueing.Queueing
	if config.QueueProvider == queueing.PROVIDER_RABBITMQ {
		addr := fmt.Sprintf(
			"%s://%s:%s@%s:%d/",
			config.QueueRabbitmqProto,
			config.QueueRabbitmqUser,
			config.QueueRabbitmqPassword,
			config.QueueRabbitmqHost,
			config.QueueRabbitmqPort,
		)
		connection := conn.NewConnection(conn.WithAddress(addr))
		que = rabbitmq.NewQueueing(rabbitmq.WithConnection(connection))
	}
	return que, nil
}
