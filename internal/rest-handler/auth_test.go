package rest_handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	rest_app "github.com/go-seidon/chariot/generated/rest-app"
	"github.com/go-seidon/chariot/internal/auth"
	mock_auth "github.com/go-seidon/chariot/internal/auth/mock"
	rest_handler "github.com/go-seidon/chariot/internal/rest-handler"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Handler", func() {
	Context("CreateClient function", Label("unit"), func() {
		var (
			currentTs   time.Time
			ctx         echo.Context
			h           func(ctx echo.Context) error
			rec         *httptest.ResponseRecorder
			authClient  *mock_auth.MockAuthClient
			createParam auth.CreateClientParam
			createRes   *auth.CreateClientResult
		)

		BeforeEach(func() {
			currentTs = time.Now()
			reqBody := &rest_app.CreateAuthClientRequest{
				ClientId:     "client-id",
				ClientSecret: "client-secret",
				Name:         "name",
				Type:         "basic",
				Status:       "active",
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
			authClient = mock_auth.NewMockAuthClient(ctrl)
			authHandler := rest_handler.NewAuth(rest_handler.AuthParam{
				AuthClient: authClient,
			})
			h = authHandler.CreateClient
			createParam = auth.CreateClientParam{
				ClientId:     reqBody.ClientId,
				ClientSecret: reqBody.ClientSecret,
				Name:         reqBody.Name,
				Type:         string(reqBody.Type),
				Status:       string(reqBody.Status),
			}
			createRes = &auth.CreateClientResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success create auth client",
				},
				Id:        "id",
				ClientId:  "client-id",
				Name:      "name",
				Type:      "basic",
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

		When("there are invalid data", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					CreateClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
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

		When("failed create client", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					CreateClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
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

		When("success create client", func() {
			It("should return result", func() {
				authClient.
					EXPECT().
					CreateClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				err := h(ctx)

				res := &rest_app.CreateAuthClientResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success create auth client"))
				Expect(res.Data).To(Equal(rest_app.CreateAuthClientData{
					Id:        createRes.Id,
					Name:      createRes.Name,
					Status:    createRes.Status,
					Type:      createRes.Type,
					ClientId:  createRes.ClientId,
					CreatedAt: createRes.CreatedAt.UnixMilli(),
				}))
			})
		})
	})

	Context("GetClientById function", Label("unit"), func() {
		var (
			currentTs  time.Time
			ctx        echo.Context
			h          func(ctx echo.Context) error
			rec        *httptest.ResponseRecorder
			authClient *mock_auth.MockAuthClient
			findParam  auth.FindClientByIdParam
			findRes    *auth.FindClientByIdResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			e := echo.New()
			ctx = e.NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("mock-id")

			t := GinkgoT()
			ctrl := gomock.NewController(t)
			authClient = mock_auth.NewMockAuthClient(ctrl)
			authHandler := rest_handler.NewAuth(rest_handler.AuthParam{
				AuthClient: authClient,
			})
			h = authHandler.GetClientById
			findParam = auth.FindClientByIdParam{
				Id: "mock-id",
			}
			findRes = &auth.FindClientByIdResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success find auth client",
				},
				Id:        "id",
				ClientId:  "client-id",
				Name:      "name",
				Type:      "basic",
				Status:    "active",
				CreatedAt: currentTs,
				UpdatedAt: &currentTs,
			}
		})

		When("there are invalid data", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					FindClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
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

		When("failed find client", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					FindClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
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

		When("client is not available", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					FindClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(nil, &system.SystemError{
						Code:    1004,
						Message: "not found",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 404,
					Message: &rest_app.ResponseBodyInfo{
						Code:    1004,
						Message: "not found",
					},
				}))
			})
		})

		When("success find client", func() {
			It("should return result", func() {
				authClient.
					EXPECT().
					FindClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				err := h(ctx)

				updatedAt := findRes.UpdatedAt.UnixMilli()

				res := &rest_app.GetAuthClientByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success find auth client"))
				Expect(res.Data).To(Equal(rest_app.GetAuthClientByIdData{
					Id:        findRes.Id,
					Name:      findRes.Name,
					Status:    findRes.Status,
					Type:      findRes.Type,
					ClientId:  findRes.ClientId,
					CreatedAt: findRes.CreatedAt.UnixMilli(),
					UpdatedAt: &updatedAt,
				}))
			})
		})
	})
})
