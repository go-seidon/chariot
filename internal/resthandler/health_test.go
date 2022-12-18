package resthandler_test

import (
	encoding_json "encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/healthcheck"
	mock_healthcheck "github.com/go-seidon/chariot/internal/healthcheck/mock"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Health Handler", func() {
	Context("CheckHealth function", Label("unit"), func() {
		var (
			ctx          echo.Context
			currentTs    time.Time
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			healthClient *mock_healthcheck.MockHealthCheck
		)

		BeforeEach(func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			e := echo.New()
			ctx = e.NewContext(req, rec)
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			healthClient = mock_healthcheck.NewMockHealthCheck(ctrl)
			healthHandler := resthandler.NewHealth(resthandler.HealthParam{
				HealthClient: healthClient,
			})
			h = healthHandler.CheckHealth
		})

		When("failed check health", func() {
			It("should return error", func() {
				healthClient.
					EXPECT().
					Check(gomock.Eq(ctx.Request().Context())).
					Return(nil, &system.Error{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 500,
					Message: &restapp.ResponseBodyInfo{
						Code:    1001,
						Message: "network error",
					},
				}))
			})
		})

		When("success check health", func() {
			It("should return result", func() {
				checkRes := &healthcheck.CheckResult{
					Success: system.Success{
						Code:    1000,
						Message: "success check health",
					},
					Status: "OK",
					Items: map[string]healthcheck.CheckResultItem{
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
					Check(gomock.Eq(ctx.Request().Context())).
					Return(checkRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.CheckHealthResponse{}
				encoding_json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success check health"))
				Expect(res.Data).To(Equal(restapp.CheckHealthData{
					Details: restapp.CheckHealthData_Details{
						AdditionalProperties: map[string]restapp.CheckHealthDetail{
							"inet": {
								Name:      "inet",
								Status:    "OK",
								Error:     "",
								CheckedAt: currentTs.UnixMilli(),
							},
						},
					},
					Status: "OK",
				}))
			})
		})

		When("success check health with no details", func() {
			It("should return result", func() {
				checkRes := &healthcheck.CheckResult{
					Success: system.Success{
						Code:    1000,
						Message: "success check health",
					},
					Status: "OK",
					Items:  map[string]healthcheck.CheckResultItem{},
				}
				healthClient.
					EXPECT().
					Check(gomock.Eq(ctx.Request().Context())).
					Return(checkRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.CheckHealthResponse{}
				encoding_json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success check health"))
				Expect(res.Data).To(Equal(restapp.CheckHealthData{
					Details: restapp.CheckHealthData_Details{
						AdditionalProperties: nil,
					},
					Status: "OK",
				}))
			})
		})
	})
})
