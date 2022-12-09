package rabbitmq

import (
	"context"
)

// @note: call init on app startup
func (que *rabbitQueue) Init(ctx context.Context) error {
	return que.conn.Init(ctx)
}
