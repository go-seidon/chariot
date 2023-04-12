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

var _ = Describe("Session Handler", func() {
	Context("CreateSession function", Label("unit"), func() {
		var (
			currentTs     time.Time
			ctx           echo.Context
			h             func(ctx echo.Context) error
			rec           *httptest.ResponseRecorder
			sessionClient *mock_service.MockSession
			createParam   service.CreateSessionParam
			createRes     *service.CreateSessionResult
		)

		BeforeEach(func() {
			currentTs = time.Now()
			reqBody := &restapp.CreateSessionRequest{
				Duration: 10,
				Features: []restapp.CreateSessionRequestFeatures{"upload_file", "retrieve_file"},
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
			sessionClient = mock_service.NewMockSession(ctrl)
			sessionHandler := resthandler.NewSession(resthandler.SessionParam{
				Session: sessionClient,
			})
			h = sessionHandler.CreateSession
			createParam = service.CreateSessionParam{
				Duration: time.Duration(reqBody.Duration),
				Features: []string{"upload_file", "retrieve_file"},
			}
			createRes = &service.CreateSessionResult{
				Success: system.Success{
					Code:    1000,
					Message: "success create session",
				},
				Token:     "token",
				ExpiresAt: currentTs,
				CreatedAt: currentTs,
			}
		})

		When("failed binding request body", func() {
			It("should return error", func() {
				body, _ := json.Marshal(struct {
					Duration string `json:"duration"`
				}{
					Duration: "x",
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
				sessionClient.
					EXPECT().
					CreateSession(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
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

		When("failed create session", func() {
			It("should return error", func() {
				sessionClient.
					EXPECT().
					CreateSession(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
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

		When("success create session", func() {
			It("should return result", func() {
				sessionClient.
					EXPECT().
					CreateSession(gomock.Eq(ctx.Request().Context()), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				err := h(ctx)

				res := &restapp.CreateSessionResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success create session"))
				Expect(res.Data).To(Equal(restapp.CreateSessionData{
					Token:     createRes.Token,
					CreatedAt: createRes.CreatedAt.UnixMilli(),
					ExpiresAt: createRes.ExpiresAt.UnixMilli(),
				}))
			})
		})
	})
})
