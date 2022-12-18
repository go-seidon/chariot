package healthcheck_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-seidon/chariot/internal/healthcheck"
	"github.com/go-seidon/provider/health"
	mock_health "github.com/go-seidon/provider/health/mock"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHealthCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Health Check Package")
}

var _ = Describe("Health Check", func() {
	Context("Check function", Label("unit"), func() {
		var (
			hc           healthcheck.HealthCheck
			healthClient *mock_health.MockHealthCheck
			ctx          context.Context
			currentTs    time.Time
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			healthClient = mock_health.NewMockHealthCheck(ctrl)
			hc = healthcheck.NewHealthCheck(healthcheck.HealthCheckParam{
				HealthClient: healthClient,
			})
		})

		When("failed check health", func() {
			It("should return error", func() {
				healthClient.
					EXPECT().
					Check(gomock.Eq(ctx)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := hc.Check(ctx)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(&system.Error{
					Code:    1001,
					Message: "network error",
				}))
			})
		})

		When("success check health", func() {
			It("should return result", func() {
				checkRes := &health.CheckResult{
					Status: "OK",
					Items: map[string]health.CheckResultItem{
						"inet": {
							Name:      "inet",
							Status:    "OK",
							Error:     "",
							Fatal:     false,
							CheckedAt: currentTs,
						},
					},
				}
				healthClient.
					EXPECT().
					Check(gomock.Eq(ctx)).
					Return(checkRes, nil).
					Times(1)

				res, err := hc.Check(ctx)

				Expect(err).To(BeNil())
				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success check health"))
				Expect(res.Status).To(Equal("OK"))
				Expect(res.Items).To(Equal(map[string]healthcheck.CheckResultItem{
					"inet": {
						Name:      "inet",
						Status:    "OK",
						Error:     "",
						Fatal:     false,
						CheckedAt: currentTs,
					},
				}))
			})
		})
	})
})
