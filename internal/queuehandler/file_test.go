package queuehandler_test

import (
	"context"
	"fmt"
	"time"

	"github.com/go-seidon/chariot/internal/file"
	mock_file "github.com/go-seidon/chariot/internal/file/mock"
	"github.com/go-seidon/chariot/internal/queuehandler"
	"github.com/go-seidon/provider/queueing"
	mock_queueing "github.com/go-seidon/provider/queueing/mock"
	mock_serialization "github.com/go-seidon/provider/serialization/mock"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/typeconv"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("File Handler", func() {

	Context("ProceedReplication function", Label("unit"), func() {
		var (
			ctx        context.Context
			currentTs  time.Time
			h          queueing.Listener
			serializer *mock_serialization.MockSerializer
			fileClient *mock_file.MockFile
			message    *mock_queueing.MockMessage
			replRes    *file.ProceedReplicationResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			fileClient = mock_file.NewMockFile(ctrl)
			message = mock_queueing.NewMockMessage(ctrl)
			fileHandler := queuehandler.NewFile(queuehandler.FileParam{
				Serializer: serializer,
				File:       fileClient,
			})
			h = fileHandler.ProceedReplication
			replRes = &file.ProceedReplicationResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success replicate file",
				},
				LocationId: typeconv.String("lid"),
				BarrelId:   typeconv.String("bid"),
				ExternalId: typeconv.String("eid"),
				UploadedAt: typeconv.Time(currentTs),
			}
		})

		When("failed drop message during unmarshal failure", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(fmt.Errorf("invalid body")).
					Times(1)

				message.
					EXPECT().
					Drop().
					Return(fmt.Errorf("drop error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("drop error")))
			})
		})

		When("failed unmarshal message", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(fmt.Errorf("invalid body")).
					Times(1)

				message.
					EXPECT().
					Drop().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("invalid body")))
			})
		})

		When("failed nack message during failure processing", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedReplication(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				message.
					EXPECT().
					Nack().
					Return(fmt.Errorf("nack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("nack error")))
			})
		})

		When("failed ack message during failure processing", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedReplication(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1004,
						Message: "file is not found",
					}).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(fmt.Errorf("ack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("ack error")))
			})
		})

		When("failure processing replication", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedReplication(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				message.
					EXPECT().
					Nack().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(&system.SystemError{
					Code:    1001,
					Message: "network error",
				}))
			})
		})

		When("failed ack message during success processing", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedReplication(gomock.Eq(ctx), gomock.Any()).
					Return(replRes, nil).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(fmt.Errorf("ack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("ack error")))
			})
		})

		When("success processing replication", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedReplication(gomock.Eq(ctx), gomock.Any()).
					Return(replRes, nil).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(BeNil())
			})
		})
	})

	Context("ProceedDeletion function", Label("unit"), func() {
		var (
			ctx context.Context
			// currentTs  time.Time
			h          queueing.Listener
			serializer *mock_serialization.MockSerializer
			fileClient *mock_file.MockFile
			message    *mock_queueing.MockMessage
			// delRes    *file.ProceedDeletionResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			// currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			fileClient = mock_file.NewMockFile(ctrl)
			message = mock_queueing.NewMockMessage(ctrl)
			fileHandler := queuehandler.NewFile(queuehandler.FileParam{
				Serializer: serializer,
				File:       fileClient,
			})
			h = fileHandler.ProceedDeletion
			// delRes = &file.ProceedDeletionResult{
			// 	Success: system.SystemSuccess{
			// 		Code:    1000,
			// 		Message: "success replicate file",
			// 	},
			// 	LocationId: typeconv.String("lid"),
			// 	BarrelId:   typeconv.String("bid"),
			// 	ExternalId: typeconv.String("eid"),
			// 	UploadedAt: typeconv.Time(currentTs),
			// }
		})

		When("failed drop message during unmarshal failure", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(fmt.Errorf("invalid body")).
					Times(1)

				message.
					EXPECT().
					Drop().
					Return(fmt.Errorf("drop error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("drop error")))
			})
		})

		When("failed unmarshal message", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(fmt.Errorf("invalid body")).
					Times(1)

				message.
					EXPECT().
					Drop().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("invalid body")))
			})
		})

		When("failed nack message during failure processing", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				message.
					EXPECT().
					Nack().
					Return(fmt.Errorf("nack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("nack error")))
			})
		})

		When("failed ack message during not found", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1004,
						Message: "file is not available",
					}).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(fmt.Errorf("ack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("ack error")))
			})
		})

		When("failed ack message during forbidden", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1003,
						Message: "deletion is already proceeded",
					}).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(fmt.Errorf("ack error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("ack error")))
			})
		})

		When("success nack message during failure processing", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, &system.SystemError{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				message.
					EXPECT().
					Nack().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(&system.SystemError{
					Code:    1001,
					Message: "network error",
				}))
			})
		})

		When("failed ack message during success deletion", func() {
			It("should return error", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, nil).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(fmt.Errorf("network error")).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("success ack message during success deletion", func() {
			It("should return result", func() {
				message.
					EXPECT().
					GetBody().
					Return([]byte{}).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Eq([]byte{}), gomock.Any()).
					Return(nil).
					Times(1)

				fileClient.
					EXPECT().
					ProceedDeletion(gomock.Eq(ctx), gomock.Any()).
					Return(nil, nil).
					Times(1)

				message.
					EXPECT().
					Ack().
					Return(nil).
					Times(1)

				err := h(ctx, message)

				Expect(err).To(BeNil())
			})
		})
	})
})
