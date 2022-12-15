package queuehandler

import (
	"context"

	"github.com/go-seidon/chariot/api/queue"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
)

type fileHandler struct {
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

	_, repErr := h.fileClient.ProceedReplication(ctx, file.ProceedReplicationParam{
		LocationId: data.LocationId,
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

	return nil
}

func (h *fileHandler) ProceedDeletion(ctx context.Context, message queueing.Message) error {
	var data queue.DeleteFileMessage
	err := h.serializer.Unmarshal(message.GetBody(), &data)
	if err != nil {
		ackErr := message.Drop()
		if ackErr != nil {
			return ackErr
		}
		return err
	}

	_, delErr := h.fileClient.ProceedDeletion(ctx, file.ProceedDeletionParam{
		LocationId: data.LocationId,
	})
	if delErr != nil {
		var ackErr error

		switch delErr.Code {
		case status.RESOURCE_NOTFOUND,
			status.ACTION_FORBIDDEN:
			ackErr = message.Ack()
		default:
			ackErr = message.Nack()
		}

		if ackErr != nil {
			return ackErr
		}
		return delErr
	}

	ackErr := message.Ack()
	if ackErr != nil {
		return ackErr
	}

	return nil
}

type FileParam struct {
	Serializer serialization.Serializer
	File       file.File
}

func NewFile(p FileParam) *fileHandler {
	return &fileHandler{
		fileClient: p.File,
		serializer: p.Serializer,
	}
}
