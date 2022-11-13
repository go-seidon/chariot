package rest_handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	rest_app "github.com/go-seidon/chariot/generated/rest-app"
	"github.com/go-seidon/chariot/internal/barrel"
	mock_barrel "github.com/go-seidon/chariot/internal/barrel/mock"
	rest_handler "github.com/go-seidon/chariot/internal/rest-handler"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Handler", func() {

	Context("CreateBarrel function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_barrel.MockBarrel
			createParam  barrel.CreateBarrelParam
			createRes    *barrel.CreateBarrelResult
		)

		BeforeEach(func() {
			currentTs = time.Now()
			reqBody := &rest_app.CreateBarrelRequest{
				Code:     "code",
				Name:     "name",
				Provider: "goseidon_hipo",
				Status:   "active",
			}
			body, _ := json.Marshal(reqBody)
			buffer := bytes.NewBuffer(body)
			req := httptest.NewRequest(http.MethodPost, "/", buffer)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			e := echo.New()
			ctx = e.NewContext(req, rec)

			t := GinkgoT()
			ctrl := gomock.NewController(t)
			barrelClient = mock_barrel.NewMockBarrel(ctrl)
			barrelHandler := rest_handler.NewBarrel(rest_handler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.CreateBarrel
			createParam = barrel.CreateBarrelParam{
				Code:     reqBody.Code,
				Name:     reqBody.Name,
				Provider: string(reqBody.Provider),
				Status:   string(reqBody.Status),
			}
			createRes = &barrel.CreateBarrelResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success create barrel",
				},
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: currentTs,
			}
		})

		When("failed binding request body", func() {
			It("should return error", func() {
				body, _ := json.Marshal(struct {
					Name int `json:"name"`
				}{
					Name: 1,
				})
				buffer := bytes.NewBuffer(body)

				req := httptest.NewRequest(http.MethodPost, "/", buffer)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				e := echo.New()
				ctx := e.NewContext(req, rec)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 400,
					Message: &rest_app.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid request",
					},
				}))
			})
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(nil, &system.SystemError{
						Code:    1002,
						Message: "invalid data",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 400,
					Message: &rest_app.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid data",
					},
				}))
			})
		})

		When("failed create barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(nil, &system.SystemError{
						Code:    1001,
						Message: "network error",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 500,
					Message: &rest_app.ResponseBodyInfo{
						Code:    1001,
						Message: "network error",
					},
				}))
			})
		})

		When("success create barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				err := h(ctx)

				res := &rest_app.CreateBarrelResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success create barrel"))
				Expect(res.Data).To(Equal(rest_app.CreateBarrelData{
					Id:        createRes.Id,
					Code:      createRes.Code,
					Name:      createRes.Name,
					Status:    createRes.Status,
					Provider:  createRes.Provider,
					CreatedAt: createRes.CreatedAt.UnixMilli(),
				}))
			})
		})
	})
})
