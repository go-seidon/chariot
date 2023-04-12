package resthandler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/chariot/internal/service"
	mock_service "github.com/go-seidon/chariot/internal/service/mock"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Barrel Handler", func() {
	Context("CreateBarrel function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_service.MockBarrel
			createParam  service.CreateBarrelParam
			createRes    *service.CreateBarrelResult
		)

		BeforeEach(func() {
			currentTs = time.Now()
			reqBody := &restapp.CreateBarrelRequest{
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
			barrelClient = mock_service.NewMockBarrel(ctrl)
			barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.CreateBarrel
			createParam = service.CreateBarrelParam{
				Code:     reqBody.Code,
				Name:     reqBody.Name,
				Provider: string(reqBody.Provider),
				Status:   string(reqBody.Status),
			}
			createRes = &service.CreateBarrelResult{
				Success: system.Success{
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
					Message: &restapp.ResponseBodyInfo{
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
					Return(nil, &system.Error{
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

		When("failed create barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
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

		When("success create barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.CreateBarrelResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success create barrel"))
				Expect(res.Data).To(Equal(restapp.CreateBarrelData{
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

	Context("GetBarrelById function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_service.MockBarrel
			findParam    service.FindBarrelByIdParam
			findRes      *service.FindBarrelByIdResult
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
			barrelClient = mock_service.NewMockBarrel(ctrl)
			barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.GetBarrelById
			findParam = service.FindBarrelByIdParam{
				Id: "mock-id",
			}
			findRes = &service.FindBarrelByIdResult{
				Success: system.Success{
					Code:    1000,
					Message: "success find barrel",
				},
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: currentTs,
				UpdatedAt: &currentTs,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(nil, &system.Error{
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

		When("failed find barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
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

		When("barrel is not available", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(nil, &system.Error{
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

		When("success find barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				err := h(ctx)

				updatedAt := findRes.UpdatedAt.UnixMilli()

				res := &restapp.GetBarrelByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success find barrel"))
				Expect(res.Data).To(Equal(restapp.GetBarrelByIdData{
					Id:        findRes.Id,
					Code:      findRes.Code,
					Name:      findRes.Name,
					Status:    findRes.Status,
					Provider:  findRes.Provider,
					CreatedAt: findRes.CreatedAt.UnixMilli(),
					UpdatedAt: &updatedAt,
				}))
			})
		})
	})

	Context("UpdateBarrelById function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_service.MockBarrel
			updateParam  service.UpdateBarrelByIdParam
			updateRes    *service.UpdateBarrelByIdResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()
			reqBody := &restapp.UpdateBarrelByIdRequest{
				Code:     "code",
				Name:     "name",
				Provider: "goseidon_hippo",
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
			barrelClient = mock_service.NewMockBarrel(ctrl)
			barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.UpdateBarrelById
			updateParam = service.UpdateBarrelByIdParam{
				Id:       "mock-id",
				Code:     reqBody.Code,
				Name:     reqBody.Name,
				Provider: string(reqBody.Provider),
				Status:   string(reqBody.Status),
			}
			updateRes = &service.UpdateBarrelByIdResult{
				Success: system.Success{
					Code:    1000,
					Message: "success update barrel",
				},
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
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
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.Error{
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

		When("barrel is not available", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.Error{
						Code:    1004,
						Message: "barrel is not available",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 404,
					Message: &restapp.ResponseBodyInfo{
						Code:    1004,
						Message: "barrel is not available",
					},
				}))
			})
		})

		When("failed update barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
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

		When("success update barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.UpdateBarrelByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success update barrel"))
				Expect(res.Data).To(Equal(restapp.UpdateBarrelByIdData{
					Id:        updateRes.Id,
					Code:      updateRes.Code,
					Name:      updateRes.Name,
					Status:    updateRes.Status,
					Provider:  updateRes.Provider,
					CreatedAt: updateRes.CreatedAt.UnixMilli(),
					UpdatedAt: updateRes.UpdatedAt.UnixMilli(),
				}))
			})
		})
	})

	Context("SearchBarrel function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_service.MockBarrel
			searchParam  service.SearchBarrelParam
			searchRes    *service.SearchBarrelResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()
			keyword := "goseidon"
			reqBody := &restapp.SearchBarrelRequest{
				Filter: &restapp.SearchBarrelFilter{
					StatusIn:   &[]restapp.SearchBarrelFilterStatusIn{"active"},
					ProviderIn: &[]restapp.SearchBarrelFilterProviderIn{"goseidon_hippo"},
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
			barrelClient = mock_service.NewMockBarrel(ctrl)
			barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.SearchBarrel
			searchParam = service.SearchBarrelParam{
				Keyword:    "goseidon",
				TotalItems: 24,
				Page:       2,
				Statuses:   []string{"active"},
				Providers:  []string{"goseidon_hippo"},
			}
			searchRes = &service.SearchBarrelResult{
				Success: system.Success{
					Code:    1000,
					Message: "success search barrel",
				},
				Items: []service.SearchBarrelItem{
					{
						Id:        "id-1",
						Code:      "code-1",
						Name:      "name-1",
						Provider:  "goseidon_hippo",
						Status:    "inactive",
						CreatedAt: currentTs,
						UpdatedAt: nil,
					},
					{
						Id:        "id-2",
						Code:      "code-2",
						Name:      "name-2",
						Provider:  "goseidon_hippo",
						Status:    "active",
						CreatedAt: currentTs,
						UpdatedAt: &currentTs,
					},
				},
				Summary: service.SearchBarrelSummary{
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
				barrelClient.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(nil, &system.Error{
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

		When("failed search barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
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

		When("there is no barrel", func() {
			It("should return empty result", func() {
				searchRes := &service.SearchBarrelResult{
					Success: system.Success{
						Code:    1000,
						Message: "success search barrel",
					},
					Items: []service.SearchBarrelItem{},
					Summary: service.SearchBarrelSummary{
						TotalItems: 0,
						Page:       2,
					},
				}
				barrelClient.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchBarrelResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search barrel"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchBarrelSummary{
					Page:       2,
					TotalItems: 0,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchBarrelItem{}))
			})
		})

		When("there is one barrel", func() {
			It("should return result", func() {
				searchRes := &service.SearchBarrelResult{
					Success: system.Success{
						Code:    1000,
						Message: "success search barrel",
					},
					Items: []service.SearchBarrelItem{
						{
							Id:        "id-1",
							Code:      "code-1",
							Name:      "name-1",
							Provider:  "goseidon_hippo",
							Status:    "active",
							CreatedAt: currentTs,
							UpdatedAt: nil,
						},
					},
					Summary: service.SearchBarrelSummary{
						TotalItems: 1,
						Page:       2,
					},
				}
				barrelClient.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchBarrelResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search barrel"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchBarrelSummary{
					Page:       2,
					TotalItems: 1,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchBarrelItem{
					{
						Id:        "id-1",
						Code:      "code-1",
						Name:      "name-1",
						Provider:  "goseidon_hippo",
						Status:    "active",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: nil,
					},
				}))
			})
		})

		When("there are some barrels", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx.Request().Context()), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.SearchBarrelResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				updatedAt := currentTs.UnixMilli()
				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success search barrel"))
				Expect(res.Data.Summary).To(Equal(restapp.SearchBarrelSummary{
					Page:       2,
					TotalItems: 2,
				}))
				Expect(res.Data.Items).To(Equal([]restapp.SearchBarrelItem{
					{
						Id:        "id-1",
						Code:      "code-1",
						Name:      "name-1",
						Provider:  "goseidon_hippo",
						Status:    "inactive",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: nil,
					},
					{
						Id:        "id-2",
						Code:      "code-2",
						Name:      "name-2",
						Provider:  "goseidon_hippo",
						Status:    "active",
						CreatedAt: currentTs.UnixMilli(),
						UpdatedAt: &updatedAt,
					},
				}))
			})
		})
	})
})
