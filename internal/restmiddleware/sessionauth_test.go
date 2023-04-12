package restmiddleware_test

import (
	"bytes"
	"mime/multipart"
	"net/http"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/restmiddleware"
	"github.com/go-seidon/chariot/internal/service"
	mock_service "github.com/go-seidon/chariot/internal/service/mock"
	mock_http "github.com/go-seidon/provider/http/mock"
	mock_serialization "github.com/go-seidon/provider/serialization/mock"
	"github.com/go-seidon/provider/system"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Session Auth Middleware", func() {
	Context("Handle Function", Label("unit"), func() {
		var (
			sessionClient *mock_service.MockSession
			serializer    *mock_serialization.MockSerializer
			handler       *mock_http.MockHandler
			m             http.Handler

			rw  *mock_http.MockResponseWriter
			req *http.Request

			verifyParam service.VerifySessionParam
			verifyRes   *service.VerifySessionResult
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			sessionClient = mock_service.NewMockSession(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			handler = mock_http.NewMockHandler(ctrl)
			fn := restmiddleware.NewSessionAuth(restmiddleware.SessionAuthParam{
				SessionClient: sessionClient,
				Serializer:    serializer,
				Feature:       "upload_file",
			})
			m = fn.Handle(handler)

			rw = mock_http.NewMockResponseWriter(ctrl)
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			err := writer.WriteField("token", "session-token")
			if err != nil {
				AbortSuite("failed create form: " + err.Error())
			}
			writer.Close()

			req, _ = http.NewRequest(http.MethodPost, "/", body)
			req.Header.Add("Content-Type", writer.FormDataContentType())

			verifyParam = service.VerifySessionParam{
				Token:   "session-token",
				Feature: "upload_file",
			}
			verifyRes = &service.VerifySessionResult{}
		})

		When("token is not specified", func() {
			It("should return error", func() {
				req, _ := http.NewRequest(http.MethodPost, "/", nil)

				b := &restapp.ResponseBodyInfo{
					Code:    1003,
					Message: "token is not specified",
				}
				serializer.
					EXPECT().
					Marshal(gomock.Eq(b)).
					Return([]byte{}, nil).
					Times(1)
				rw.
					EXPECT().
					Header().
					Return(map[string][]string{}).
					Times(1)
				rw.
					EXPECT().
					WriteHeader(403).
					Times(1)
				rw.
					EXPECT().
					Write(gomock.Eq([]byte{})).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("failed check session", func() {
			It("should return error", func() {
				sessionClient.
					EXPECT().
					VerifySession(gomock.Eq(req.Context()), gomock.Eq(verifyParam)).
					Return(nil, &system.Error{
						Code:    1001,
						Message: "disk error",
					}).
					Times(1)

				b := &restapp.ResponseBodyInfo{
					Code:    1003,
					Message: "disk error",
				}
				serializer.
					EXPECT().
					Marshal(gomock.Eq(b)).
					Return([]byte{}, nil).
					Times(1)
				rw.
					EXPECT().
					Header().
					Return(map[string][]string{}).
					Times(1)
				rw.
					EXPECT().
					WriteHeader(403).
					Times(1)
				rw.
					EXPECT().
					Write(gomock.Eq([]byte{})).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("session is valid", func() {
			It("should return result", func() {
				sessionClient.
					EXPECT().
					VerifySession(gomock.Eq(req.Context()), gomock.Eq(verifyParam)).
					Return(verifyRes, nil).
					Times(1)

				handler.
					EXPECT().
					ServeHTTP(gomock.Eq(rw), gomock.Eq(req)).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("token is specified through query param", func() {
			It("should return result", func() {
				req, _ := http.NewRequest(http.MethodPost, "/?token=session-token", nil)

				sessionClient.
					EXPECT().
					VerifySession(gomock.Eq(req.Context()), gomock.Eq(verifyParam)).
					Return(verifyRes, nil).
					Times(1)

				handler.
					EXPECT().
					ServeHTTP(gomock.Eq(rw), gomock.Eq(req)).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})
	})
})
