package queuehandler

import (
	"context"
	"time"

	"github.com/go-seidon/chariot/api/queue"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
)

type fileHandler struct {
	logger     logging.Logger
	serializer serialization.Serializer
	fileClient file.File
}

func (h *fileHandler) ProceedReplication(ctx context.Context, message queueing.Message) error {
	var data queue.ScheduleReplicationMessage
	err := h.serializer.Unmarshal(message.GetBody(), &data)
	if err != nil {
		ackErr := message.Drop()
		if ackErr != nil {
			return ackErr
		}
		return err
	}

	proceeded, repErr := h.fileClient.ProceedReplication(ctx, file.ProceedReplicationParam{
		LocationId: data.Id,
	})
	if repErr != nil {
		var ackErr error

		switch repErr.Code {
		case status.RESOURCE_NOTFOUND:
			ackErr = message.Ack()
		default:
			ackErr = message.Nack()
		}

		if ackErr != nil {
			return ackErr
		}
		return repErr
	}

	ackErr := message.Ack()
	if ackErr != nil {
		return ackErr
	}

	h.logger.WithFields(map[string]interface{}{
		"started_at":  proceeded.StartedAt.Format(time.RFC3339),
		"finished_at": proceeded.ProceededAt.Format(time.RFC3339),
	}).Infof(proceeded.Success.Message)

	return nil
}

type FileParam struct {
	Logger     logging.Logger
	Serializer serialization.Serializer
	File       file.File
}

func NewFile(p FileParam) *fileHandler {
	return &fileHandler{
		logger:     p.Logger,
		fileClient: p.File,
		serializer: p.Serializer,
	}
}
