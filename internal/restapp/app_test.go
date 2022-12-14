package restapp_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/restapp"

	mock_queue "github.com/go-seidon/chariot/internal/queue/mock"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	mock_restapp "github.com/go-seidon/chariot/internal/restapp/mock"
	mock_logging "github.com/go-seidon/provider/logging/mock"
	mock_queueing "github.com/go-seidon/provider/queueing/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rest App Package")
}

var _ = Describe("App Package", func() {

	Context("NewRestApp function", Label("unit"), func() {
		When("config is not specified", func() {
			It("should return error", func() {
				res, err := restapp.NewRestApp()

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("required parameters are specified", func() {
			It("should return error", func() {
				config := &app.Config{
					RepositoryProvider: "mysql",
					QueueProvider:      "rabbitmq",
				}
				res, err := restapp.NewRestApp(
					restapp.WithConfig(config),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("logger is specified", func() {
			It("should return result", func() {
				t := GinkgoT()
				ctrl := gomock.NewController(t)
				config := &app.Config{
					RepositoryProvider: "mysql",
					QueueProvider:      "rabbitmq",
				}
				logger := mock_logging.NewMockLogger(ctrl)
				res, err := restapp.NewRestApp(
					restapp.WithConfig(config),
					restapp.WithLogger(logger),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("repository is specified", func() {
			It("should return result", func() {
				t := GinkgoT()
				ctrl := gomock.NewController(t)
				config := &app.Config{
					RepositoryProvider: "mysql",
					QueueProvider:      "rabbitmq",
				}
				repo := mock_repository.NewMockProvider(ctrl)
				repo.
					EXPECT().
					GetAuth().
					Times(2)

				repo.
					EXPECT().
					GetBarrel().
					Times(2)

				repo.
					EXPECT().
					GetFile().
					Times(1)

				res, err := restapp.NewRestApp(
					restapp.WithConfig(config),
					restapp.WithRepository(repo),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("server is specified", func() {
			It("should return result", func() {
				t := GinkgoT()
				ctrl := gomock.NewController(t)
				config := &app.Config{
					RepositoryProvider: "mysql",
					QueueProvider:      "rabbitmq",
				}
				server := mock_restapp.NewMockServer(ctrl)
				res, err := restapp.NewRestApp(
					restapp.WithConfig(config),
					restapp.WithServer(server),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Context("Start function", Label("unit"), func() {
		var (
			rApp   app.App
			ctx    context.Context
			server *mock_restapp.MockServer
			logger *mock_logging.MockLogger
			repo   *mock_repository.MockProvider
			queue  *mock_queue.MockQueue
			wg     *sync.WaitGroup
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			server = mock_restapp.NewMockServer(ctrl)
			logger = mock_logging.NewMockLogger(ctrl)
			repo = mock_repository.NewMockProvider(ctrl)
			queuer := mock_queueing.NewMockQueuer(ctrl)
			queue = mock_queue.NewMockQueue(ctrl)
			config := &app.Config{
				AppName:     "name",
				AppVersion:  "version",
				RESTAppHost: "host",
				RESTAppPort: 1,
			}
			rApp, _ = restapp.NewRestApp(
				restapp.WithConfig(config),
				restapp.WithServer(server),
				restapp.WithLogger(logger),
				restapp.WithRepository(repo),
				restapp.WithQueue(queue),
				restapp.WithQueuer(queuer),
			)
			wg = &sync.WaitGroup{}
		})

		When("failed init repo", func() {
			It("should return error", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Running %s:%s"),
						gomock.Eq("name-rest"),
						gomock.Eq("version"),
					).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Initializing repository")).
					Times(1)

				repo.
					EXPECT().
					Init(gomock.Eq(ctx)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := rApp.Run(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed start queue", func() {
			It("should return error", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Running %s:%s"),
						gomock.Eq("name-rest"),
						gomock.Eq("version"),
					).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Initializing repository")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on queue")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on: %s"), gomock.Eq("host:1")).
					Times(1)

				repo.
					EXPECT().
					Init(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(2)

				queue.
					EXPECT().
					Start(gomock.Eq(ctx)).
					DoAndReturn(func(_ context.Context) error {
						wg.Done()
						return fmt.Errorf("network error")
					}).
					Times(1)

				server.
					EXPECT().
					Start(gomock.Eq("host:1")).
					DoAndReturn(func(_ string) error {
						wg.Done()
						return nil
					}).
					Times(1)

				err := rApp.Run(ctx)
				wg.Wait()

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed start server", func() {
			It("should return error", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Running %s:%s"),
						gomock.Eq("name-rest"),
						gomock.Eq("version"),
					).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Initializing repository")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on queue")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on: %s"), gomock.Eq("host:1")).
					Times(1)

				repo.
					EXPECT().
					Init(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(2)

				queue.
					EXPECT().
					Start(gomock.Eq(ctx)).
					DoAndReturn(func(_ context.Context) error {
						wg.Done()
						return nil
					}).
					Times(1)

				server.
					EXPECT().
					Start(gomock.Eq("host:1")).
					DoAndReturn(func(_ string) error {
						wg.Done()
						return fmt.Errorf("network error")
					}).
					Times(1)

				err := rApp.Run(ctx)
				wg.Wait()

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("success start server and queue", func() {
			It("should return result", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Running %s:%s"),
						gomock.Eq("name-rest"),
						gomock.Eq("version"),
					).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Initializing repository")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on queue")).
					Times(1)

				logger.
					EXPECT().
					Infof(gomock.Eq("Listening on: %s"), gomock.Eq("host:1")).
					Times(1)

				repo.
					EXPECT().
					Init(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				wg.Add(2)

				queue.
					EXPECT().
					Start(gomock.Eq(ctx)).
					DoAndReturn(func(_ context.Context) error {
						wg.Done()
						return nil
					}).
					Times(1)

				server.
					EXPECT().
					Start(gomock.Eq("host:1")).
					DoAndReturn(func(_ string) error {
						wg.Done()
						return nil
					}).
					Times(1)

				err := rApp.Run(ctx)
				wg.Wait()

				Expect(err).To(BeNil())
			})
		})
	})

	Context("Stop function", Label("unit"), func() {
		var (
			rApp   app.App
			ctx    context.Context
			server *mock_restapp.MockServer
			logger *mock_logging.MockLogger
			repo   *mock_repository.MockProvider
			queue  *mock_queue.MockQueue
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			server = mock_restapp.NewMockServer(ctrl)
			logger = mock_logging.NewMockLogger(ctrl)
			repo = mock_repository.NewMockProvider(ctrl)
			queuer := mock_queueing.NewMockQueuer(ctrl)
			queue = mock_queue.NewMockQueue(ctrl)
			config := &app.Config{
				AppName:     "name",
				AppVersion:  "version",
				RESTAppHost: "host",
				RESTAppPort: 1,
			}
			rApp, _ = restapp.NewRestApp(
				restapp.WithConfig(config),
				restapp.WithServer(server),
				restapp.WithLogger(logger),
				restapp.WithRepository(repo),
				restapp.WithQueue(queue),
				restapp.WithQueuer(queuer),
			)
		})

		When("failed stop server", func() {
			It("should return error", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Stopping %s on: %s"),
						gomock.Eq("name-rest"),
						gomock.Eq("host:1"),
					).
					Times(1)

				server.
					EXPECT().
					Shutdown(gomock.Eq(ctx)).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := rApp.Stop(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("success stop server", func() {
			It("should return result", func() {
				logger.
					EXPECT().
					Infof(
						gomock.Eq("Stopping %s on: %s"),
						gomock.Eq("name-rest"),
						gomock.Eq("host:1"),
					).
					Times(1)

				server.
					EXPECT().
					Shutdown(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				err := rApp.Stop(ctx)

				Expect(err).To(BeNil())
			})
		})
	})

})
