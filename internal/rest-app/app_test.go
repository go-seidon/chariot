package rest_app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-seidon/chariot/internal/app"
	rest_app "github.com/go-seidon/chariot/internal/rest-app"
	mock_rest_app "github.com/go-seidon/chariot/internal/rest-app/mock"
	mock_logging "github.com/go-seidon/provider/logging/mock"
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
				res, err := rest_app.NewRestApp()

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("required parameters are specified", func() {
			It("should return error", func() {
				config := &app.Config{}
				res, err := rest_app.NewRestApp(
					rest_app.WithConfig(config),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("logger is specified", func() {
			It("should return error", func() {
				t := GinkgoT()
				ctrl := gomock.NewController(t)
				config := &app.Config{}
				logger := mock_logging.NewMockLogger(ctrl)
				res, err := rest_app.NewRestApp(
					rest_app.WithConfig(config),
					rest_app.WithLogger(logger),
				)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("server is specified", func() {
			It("should return error", func() {
				t := GinkgoT()
				ctrl := gomock.NewController(t)
				config := &app.Config{}
				server := mock_rest_app.NewMockServer(ctrl)
				res, err := rest_app.NewRestApp(
					rest_app.WithConfig(config),
					rest_app.WithServer(server),
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
			server *mock_rest_app.MockServer
			logger *mock_logging.MockLogger
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			server = mock_rest_app.NewMockServer(ctrl)
			logger = mock_logging.NewMockLogger(ctrl)
			config := &app.Config{
				AppName:     "name",
				AppVersion:  "version",
				RESTAppHost: "host",
				RESTAppPort: 1,
			}
			rApp, _ = rest_app.NewRestApp(
				rest_app.WithConfig(config),
				rest_app.WithServer(server),
				rest_app.WithLogger(logger),
			)
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
					Infof(gomock.Eq("Listening on: %s"), gomock.Eq("host:1")).
					Times(1)

				server.
					EXPECT().
					Start(gomock.Eq("host:1")).
					Return(fmt.Errorf("network error")).
					Times(1)

				err := rApp.Run(ctx)

				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("success start server", func() {
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
					Infof(gomock.Eq("Listening on: %s"), gomock.Eq("host:1")).
					Times(1)

				server.
					EXPECT().
					Start(gomock.Eq("host:1")).
					Return(nil).
					Times(1)

				err := rApp.Run(ctx)

				Expect(err).To(BeNil())
			})
		})
	})

	Context("Stop function", Label("unit"), func() {
		var (
			rApp   app.App
			ctx    context.Context
			server *mock_rest_app.MockServer
			logger *mock_logging.MockLogger
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			server = mock_rest_app.NewMockServer(ctrl)
			logger = mock_logging.NewMockLogger(ctrl)
			config := &app.Config{
				AppName:     "name",
				AppVersion:  "version",
				RESTAppHost: "host",
				RESTAppPort: 1,
			}
			rApp, _ = rest_app.NewRestApp(
				rest_app.WithConfig(config),
				rest_app.WithServer(server),
				rest_app.WithLogger(logger),
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
