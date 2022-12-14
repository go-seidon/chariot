package resthandler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/auth"
	mock_auth "github.com/go-seidon/chariot/internal/auth/mock"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Handler", func() {
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
			reqBody := &restapp.CreateAuthClientRequest{
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
			authHandler := resthandler.NewAuth(resthandler.AuthParam{
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
					Message: &restapp.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid request",
					},
				}))
			})
		})

		When("there is invalid data", func() {
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
					Message: &restapp.ResponseBodyInfo{
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
					Message: &restapp.ResponseBodyInfo{
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

				res := &restapp.CreateAuthClientResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success create auth client"))
				Expect(res.Data).To(Equal(restapp.CreateAuthClientData{
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
			authHandler := resthandler.NewAuth(resthandler.AuthParam{
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

		When("there is invalid data", func() {
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
					Message: &restapp.ResponseBodyInfo{
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
					Message: &restapp.ResponseBodyInfo{
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
					Message: &restapp.ResponseBodyInfo{
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

				res := &restapp.GetAuthClientByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success find auth client"))
				Expect(res.Data).To(Equal(restapp.GetAuthClientByIdData{
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

	Context("UpdateClientById function", Label("unit"), func() {
		var (
			currentTs   time.Time
			ctx         echo.Context
			h           func(ctx echo.Context) error
			rec         *httptest.ResponseRecorder
			authClient  *mock_auth.MockAuthClient
			updateParam auth.UpdateClientByIdParam
			updateRes   *auth.UpdateClientByIdResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()
			reqBody := &restapp.UpdateAuthClientByIdRequest{
				ClientId: "client-id",
				Name:     "name",
				Type:     "basic",
				Status:   "active",
			}
			body, _ := json.Marshal(reqBody)
			buffer := bytes.NewBuffer(body)
			req := httptest.NewRequest(http.MethodPost, "/", buffer)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			e := echo.New()
			ctx = e.NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("mock-id")

			t := GinkgoT()
			ctrl := gomock.NewController(t)
			authClient = mock_auth.NewMockAuthClient(ctrl)
			authHandler := resthandler.NewAuth(resthandler.AuthParam{
				AuthClient: authClient,
			})
			h = authHandler.UpdateClientById
			updateParam = auth.UpdateClientByIdParam{
				Id:       "mock-id",
				ClientId: reqBody.ClientId,
				Name:     reqBody.Name,
				Type:     string(reqBody.Type),
				Status:   string(reqBody.Status),
			}
			updateRes = &auth.UpdateClientByIdResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success update auth client",
				},
				Id:        "id",
				ClientId:  "client-id",
				Name:      "name",
				Type:      "basic",
				Status:    "active",
				CreatedAt: currentTs,
				UpdatedAt: currentTs,
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
					Message: &restapp.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid request",
					},
				}))
			})
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					UpdateClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.SystemError{
						Code:    1002,
						Message: "invalid data",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 400,
					Message: &restapp.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid data",
					},
				}))
			})
		})

		When("auth client is not available", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					UpdateClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.SystemError{
						Code:    1004,
						Message: "auth client is not available",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 404,
					Message: &restapp.ResponseBodyInfo{
						Code:    1004,
						Message: "auth client is not available",
					},
				}))
			})
		})

		When("failed update auth client", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					UpdateClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.SystemError{
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

		When("success update auth client", func() {
			It("should return result", func() {
				authClient.
					EXPECT().
					UpdateClientById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.UpdateAuthClientByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success update auth client"))
				Expect(res.Data).To(Equal(restapp.UpdateAuthClientByIdData{
					Id:        updateRes.Id,
					Name:      updateRes.Name,
					Status:    updateRes.Status,
					Type:      updateRes.Type,
					ClientId:  updateRes.ClientId,
					CreatedAt: updateRes.CreatedAt.UnixMilli(),
					UpdatedAt: updateRes.UpdatedAt.UnixMilli(),
				}))
			})
		})
	})

	Context("SearchClient function", Label("unit"), func() {
		var (
			currentTs   time.Time
			ctx         echo.Context
			h           func(ctx echo.Context) error
			rec         *httptest.ResponseRecorder
			authClient  *mock_auth.MockAuthClient
			searchParam auth.SearchClientParam
			searchRes   *auth.SearchClientResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()
			keyword := "goseidon"
			reqBody := &restapp.SearchAuthClientRequest{
				Filter: &restapp.SearchAuthClientFilter{
					StatusIn: &[]restapp.SearchAuthClientFilterStatusIn{"active"},
				},
				Keyword: &keyword,
				Pagination: &restapp.RequestPagination{
					Page:       2,
					TotalItems: 24,
				},
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
			authHandler := resthandler.NewAuth(resthandler.AuthParam{
				AuthClient: authClient,
			})
			h = authHandler.SearchClient
			searchParam = auth.SearchClientParam{
				Keyword:    "goseidon",
				TotalItems: 24,
				Page:       2,
				Statuses:   []string{"active"},
			}
			searchRes = &auth.SearchClientResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success search auth client",
				},
				Items: []auth.SearchClientItem{
					{
						Id:        "id-1",
						ClientId:  "client-id-1",
						Name:      "name-1",
						Type:      "basic",
						Status:    "inactive",
						CreatedAt: currentTs,
						UpdatedAt: nil,
					},
					{
						Id:        "id-2",
						ClientId:  "client-id-2",
						Name:      "name-2",
						Type:      "basic",
						Status:    "active",
						CreatedAt: currentTs,
						UpdatedAt: &currentTs,
					},
				},
				Summary: auth.SearchClientSummary{
					TotalItems: 2,
					Page:       2,
				},
			}
		})

		When("failed binding request body", func() {
			It("should return error", func() {
				reqBody, _ := json.Marshal(struct {
					Filter int `json:"filter"`
				}{
					Filter: 1,
				})
				buffer := bytes.NewBuffer(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/", buffer)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				e := echo.New()
				ctx := e.NewContext(req, rec)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 400,
					Message: &restapp.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid request",
					},
				}))
			})
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					SearchClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(nil, &system.SystemError{
						Code:    1002,
						Message: "invalid data",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 400,
					Message: &restapp.ResponseBodyInfo{
						Code:    1002,
						Message: "invalid data",
					},
				}))
			})
		})

		When("failed search client", func() {
			It("should return error", func() {
				authClient.
					EXPECT().
					SearchClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(nil, &system.SystemError{
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

		When("there is no client", func() {
			It("should return empty result", func() {
				searchRes := &auth.SearchClientResult{
					Success: system.SystemSuccess{
						Code:    1000,
						Message: "success search auth client",
					},
					Items: []auth.SearchClientItem{},
					Summary: auth.SearchClientSummary{
						TotalItems: 0,
						Page:       2,
					},
				}
				authClient.
					EXPECT().
					SearchClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchAuthClientResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search auth client"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchAuthClientSummary{
					Page:       2,
					TotalItems: 0,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchAuthClientItem{}))
			})
		})

		When("there is one client", func() {
			It("should return result", func() {
				searchRes := &auth.SearchClientResult{
					Success: system.SystemSuccess{
						Code:    1000,
						Message: "success search auth client",
					},
					Items: []auth.SearchClientItem{
						{
							Id:        "id-1",
							ClientId:  "client-id-1",
							Name:      "name-1",
							Type:      "basic",
							Status:    "active",
							CreatedAt: currentTs,
							UpdatedAt: nil,
						},
					},
					Summary: auth.SearchClientSummary{
						TotalItems: 1,
						Page:       2,
					},
				}
				authClient.
					EXPECT().
					SearchClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchAuthClientResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search auth client"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchAuthClientSummary{
					Page:       2,
					TotalItems: 1,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchAuthClientItem{
					{
						Id:        "id-1",
						ClientId:  "client-id-1",
						Name:      "name-1",
						Type:      "basic",
						Status:    "active",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: nil,
					},
				}))
			})
		})

		When("there are some clients", func() {
			It("should return result", func() {
				authClient.
					EXPECT().
					SearchClient(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchAuthClientResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				updatedAt := currentTs.UnixMilli()
				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search auth client"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchAuthClientSummary{
					Page:       2,
					TotalItems: 2,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchAuthClientItem{
					{
						Id:        "id-1",
						ClientId:  "client-id-1",
						Name:      "name-1",
						Type:      "basic",
						Status:    "inactive",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: nil,
					},
					{
						Id:        "id-2",
						ClientId:  "client-id-2",
						Name:      "name-2",
						Type:      "basic",
						Status:    "active",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: &updatedAt,
					},
				}))
			})
		})
	})
})
