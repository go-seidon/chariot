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

	Context("GetBarrelById function", Label("unit"), func() {
		var (
			currentTs    time.Time
			ctx          echo.Context
			h            func(ctx echo.Context) error
			rec          *httptest.ResponseRecorder
			barrelClient *mock_barrel.MockBarrel
			findParam    barrel.FindBarrelByIdParam
			findRes      *barrel.FindBarrelByIdResult
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
			barrelClient = mock_barrel.NewMockBarrel(ctrl)
			barrelHandler := rest_handler.NewBarrel(rest_handler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.GetBarrelById
			findParam = barrel.FindBarrelByIdParam{
				Id: "mock-id",
			}
			findRes = &barrel.FindBarrelByIdResult{
				Success: system.SystemSuccess{
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

		When("failed find barrel", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
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

		When("barrel is not available", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
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

		When("success find barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					FindBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				err := h(ctx)

				updatedAt := findRes.UpdatedAt.UnixMilli()

				res := &rest_app.GetBarrelByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success find barrel"))
				Expect(res.Data).To(Equal(rest_app.GetBarrelByIdData{
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
			barrelClient *mock_barrel.MockBarrel
			updateParam  barrel.UpdateBarrelByIdParam
			updateRes    *barrel.UpdateBarrelByIdResult
		)

		BeforeEach(func() {
			currentTs = time.Now().UTC()
			reqBody := &rest_app.UpdateBarrelByIdRequest{
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
			barrelClient = mock_barrel.NewMockBarrel(ctrl)
			barrelHandler := rest_handler.NewBarrel(rest_handler.BarrelParam{
				Barrel: barrelClient,
			})
			h = barrelHandler.UpdateBarrelById
			updateParam = barrel.UpdateBarrelByIdParam{
				Id:       "mock-id",
				Code:     reqBody.Code,
				Name:     reqBody.Name,
				Provider: string(reqBody.Provider),
				Status:   string(reqBody.Status),
			}
			updateRes = &barrel.UpdateBarrelByIdResult{
				Success: system.SystemSuccess{
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
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
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

		When("barrel is not available", func() {
			It("should return error", func() {
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(nil, &system.SystemError{
						Code:    1004,
						Message: "barrel is not available",
					}).
					Times(1)

				err := h(ctx)

				Expect(err).To(Equal(&echo.HTTPError{
					Code: 404,
					Message: &rest_app.ResponseBodyInfo{
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

		When("success update barrel", func() {
			It("should return result", func() {
				barrelClient.
					EXPECT().
					UpdateBarrelById(gomock.Eq(ctx.Request().Context()), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				err := h(ctx)

				res := &rest_app.UpdateBarrelByIdResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success update barrel"))
				Expect(res.Data).To(Equal(rest_app.UpdateBarrelByIdData{
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
})
