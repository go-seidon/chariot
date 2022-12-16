package queue_test

import (
	"context"
	"fmt"
	"sync"
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
			ctx    context.Context
			que    queue.Queue
			queuer *mock_queueing.MockQueuer
			wg     *sync.WaitGroup
			// decExc1Param  queueing.DeclareExchangeParam
			// decQue1Param  queueing.DeclareQueueParam
			// decQue1Res    *queueing.DeclareQueueResult
			// bindQue1Param queueing.BindQueueParam
			// decExc2Param  queueing.DeclareExchangeParam
			// decQue2Param  queueing.DeclareQueueParam
			// decQue2Res    *queueing.DeclareQueueResult
			// bindQue2Param queueing.BindQueueParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			queuer = mock_queueing.NewMockQueuer(ctrl)
			que = queue.NewQueue(queue.QueueParam{
				Queuer: queuer,
			})
			wg = &sync.WaitGroup{}
			// decExc1Param = queueing.DeclareExchangeParam{
			// 	ExchangeName: "file_replication",
			// 	ExchangeType: "fanout",
			// }
			// decQue1Param = queueing.DeclareQueueParam{
			// 	QueueName: "proceed_file_replication",
			// }
			// decQue1Res = &queueing.DeclareQueueResult{
			// 	Name: "proceed_file_replication",
			// }
			// bindQue1Param = queueing.BindQueueParam{
			// 	ExchangeName: "file_replication",
			// 	QueueName:    "proceed_file_replication",
			// }
			// decExc2Param = queueing.DeclareExchangeParam{
			// 	ExchangeName: "file_deletion",
			// 	ExchangeType: "fanout",
			// }
			// decQue2Param = queueing.DeclareQueueParam{
			// 	QueueName: "proceed_file_deletion",
			// }
			// decQue2Res = &queueing.DeclareQueueResult{
			// 	Name: "proceed_file_deletion",
			// }
			// bindQue2Param = queueing.BindQueueParam{
			// 	ExchangeName: "file_deletion",
			// 	QueueName:    "proceed_file_deletion",
			// }
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

		When("failed declare exchange", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(4)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareExchangeParam) error {
						wg.Done()
						return fmt.Errorf("exchange error")
					}).
					Times(2)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
						wg.Done()
						return nil, nil
					}).
					Times(2)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("exchange error")))
			})
		})

		When("failed declare queue", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(4)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareExchangeParam) error {
						wg.Done()
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
						wg.Done()
						return nil, fmt.Errorf("queue error")
					}).
					Times(2)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("queue error")))
			})
		})

		When("failed bind queue", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(6)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareExchangeParam) error {
						wg.Done()
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
						wg.Done()
						return nil, nil
					}).
					Times(2)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.BindQueueParam) error {
						wg.Done()
						return fmt.Errorf("queue error")
					}).
					Times(2)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("queue error")))
			})
		})

		When("failed subscribe", func() {
			It("should return error", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(8)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareExchangeParam) error {
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
						return nil, nil
					}).
					Times(2)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.BindQueueParam) error {
						wg.Done()
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.SubscribeParam) error {
						wg.Done()
						return fmt.Errorf("subscribe error")
					}).
					Times(2)

				err := que.Start(ctx)

				Expect(err).To(Equal(fmt.Errorf("subscribe error")))
			})
		})

		When("success start queue", func() {
			It("should return result", func() {
				queuer.
					EXPECT().
					Open(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(8)

				queuer.
					EXPECT().
					DeclareExchange(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareExchangeParam) error {
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					DeclareQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
						return nil, nil
					}).
					Times(2)

				queuer.
					EXPECT().
					BindQueue(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.BindQueueParam) error {
						wg.Done()
						return nil
					}).
					Times(2)

				queuer.
					EXPECT().
					Subscribe(gomock.Eq(ctx), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ queueing.SubscribeParam) error {
						wg.Done()
						return nil
					}).
					Times(2)

				err := que.Start(ctx)

				Expect(err).To(BeNil())
			})
		})
	})
})
