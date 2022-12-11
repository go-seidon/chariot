package queuehandler

import (
	"context"
	"fmt"
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

// @todo: add implmenetation + unit test
func (h *fileHandler) ProceedReplication(ctx context.Context, message queueing.Message) error {
	h.logger.Logf(
		"Processing message: %s published at: %s",
		message.GetId(),
		message.GetPublishedAt().UTC().Format(time.RFC3339),
	)

	var data queue.ScheduleReplicationMessage
	err := h.serializer.Unmarshal(message.GetBody(), &data)
	if err != nil {
		h.logger.Errorf("Failed unmarshal message: %v", err)
		return err
	}

	fmt.Println("Id", data.Id)
	fmt.Println("Priority", data.Priority)
	fmt.Println("Status", data.Status)
	fmt.Println("BarrelId", data.BarrelId)
	fmt.Println("FileId", data.FileId)

	proceeded, ferr := h.fileClient.ProceedReplication(ctx, file.ProceedReplicationParam{
		LocationId: data.Id,
	})
	if ferr != nil {
		h.logger.Errorf("Failed proceed replication: %v", ferr)
		if ferr.Code != status.RESOURCE_NOTFOUND {
			err := message.Drop()
			if err != nil {
				h.logger.Errorf("Failed nacking message: %v", err)
				return err
			}
		}
	}

	err = message.Ack()
	if err != nil {
		h.logger.Errorf("Failed acknowledge message: %v", err)
		return err
	}

	h.logger.Infof("%s", proceeded.Success.Message)
	h.logger.Infof("started at: %s, finished at: %s", proceeded.StartedAt.Format(time.RFC3339), proceeded.ProceededAt.Format(time.RFC3339))

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
