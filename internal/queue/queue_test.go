package queue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-seidon/chariot/internal/queue"
	"github.com/go-seidon/provider/queueing"
	mock_queueing "github.com/go-seidon/provider/queueing/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue Package")
}

var _ = Describe("Queue Package", func() {
	Context("Start function", Label("unit"), func() {
		var (
			ctx           context.Context
			que           queue.Queue
			queuer        *mock_queueing.MockQueuer
			decExc1Param  queueing.DeclareExchangeParam
			decQue1Param  queueing.DeclareQueueParam
			decQue1Res    *queueing.DeclareQueueResult
			bindQue1Param queueing.BindQueueParam
			decExc2Param  queueing.DeclareExchangeParam
			decQue2Param  queueing.DeclareQueueParam
			decQue2Res    *queueing.DeclareQueueResult
			bindQue2Param queueing.BindQueueParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			queuer = mock_queueing.NewMockQueuer(ctrl)
			que = queue.NewQueue(queue.QueueParam{
				Queuer: queuer,
			})
			decExc1Param = queueing.DeclareExchangeParam{
				ExchangeName: "file_replication",
				ExchangeType: "fanout",
			}
			decQue1Param = queueing.DeclareQueueParam{
				QueueName: "proceed_file_replication",
			}
			decQue1Res = &queueing.DeclareQueueResult{
				Name: "proceed_file_replication",
			}
			bindQue1Param = queueing.BindQueueParam{
				ExchangeName: "file_replication",
				QueueName:    "proceed_file_replication",
			}
			decExc2Param = queueing.DeclareExchangeParam{
				ExchangeName: "file_deletion",
				ExchangeType: "fanout",
			}
			decQue2Param = queueing.DeclareQueueParam{
				QueueName: "proceed_file_deletion",
			}
			decQue2Res = &queueing.DeclareQueueResult{
				Name: "proceed_file_deletion",
			}
			bindQue2Param = queueing.BindQueueParam{
				ExchangeName: "file_deletion",
				QueueName:    "proceed_file_deletion",
			}
		})

		When("failed open queuer connection", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed declare exchange file_replication", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed declare queue proceed_file_replication", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed bind queue proceed_file_replication", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed subscribe proceed_file_replication", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed declare exchange file_deletion", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc2Param)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed declare queue proceed_file_deletion", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc2Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue2Param)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed bind queue proceed_file_deletion", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc2Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue2Param)).
					Return(decQue2Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue2Param)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("success start queue", func() {
			It("should return result", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue1Param)).
					Return(decQue1Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue1Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Eq(decExc2Param)).
					Return(nil).
					Times(1)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Eq(decQue2Param)).
					Return(decQue2Res, nil).
					Times(1)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Eq(bindQue2Param)).
					Return(nil).
					Times(1)

				err := que.Start(ctx)

				Expect(err).To(BeNil())
			})
		})
	})
})
