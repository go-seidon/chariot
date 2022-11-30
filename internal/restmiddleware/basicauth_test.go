package restmiddleware_test

import (
	"fmt"
	"net/http"

	"github.com/go-seidon/chariot/generated/restapp"
	"github.com/go-seidon/chariot/internal/auth"
	mock_auth "github.com/go-seidon/chariot/internal/auth/mock"
	"github.com/go-seidon/chariot/internal/restmiddleware"
	mock_http "github.com/go-seidon/provider/http/mock"
	mock_serialization "github.com/go-seidon/provider/serialization/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Basic Auth Middleware", func() {
	Context("Handle Function", Label("unit"), func() {
		var (
			a       *mock_auth.MockBasicAuth
			s       *mock_serialization.MockSerializer
			handler *mock_http.MockHandler
			m       http.Handler

			rw  *mock_http.MockResponseWriter
			req *http.Request

			checkParam auth.CheckCredentialParam
			checkRes   *auth.CheckCredentialResult
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			a = mock_auth.NewMockBasicAuth(ctrl)
			s = mock_serialization.NewMockSerializer(ctrl)
			handler = mock_http.NewMockHandler(ctrl)
			fn := restmiddleware.NewBasicAuth(restmiddleware.BasicAuthParam{
				BasicClient: a,
				Serializer:  s,
			})
			m = fn.Handle(handler)

			rw = mock_http.NewMockResponseWriter(ctrl)
			req = &http.Request{
				Header: http.Header{},
			}
			req.Header.Set("Authorization", "Basic basic-token")

			checkParam = auth.CheckCredentialParam{
				AuthToken: "basic-token",
			}
			checkRes = &auth.CheckCredentialResult{
				TokenValid: true,
			}
		})

		When("authorization header is not specified", func() {
			It("should return error", func() {
				req.Header.Del("Authorization")

				b := &restapp.ResponseBodyInfo{
					Code:    1003,
					Message: "credential is not specified",
				}
				s.
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
					WriteHeader(401).
					Times(1)
				rw.
					EXPECT().
					Write(gomock.Eq([]byte{})).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("failed check credential", func() {
			It("should return error", func() {
				a.
					EXPECT().
					CheckCredential(gomock.Eq(req.Context()), gomock.Eq(checkParam)).
					Return(nil, fmt.Errorf("db error")).
					Times(1)

				b := &restapp.ResponseBodyInfo{
					Code:    1003,
					Message: "failed check credential",
				}
				s.
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
					WriteHeader(401).
					Times(1)
				rw.
					EXPECT().
					Write(gomock.Eq([]byte{})).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("failed token is invalid", func() {
			It("should return error", func() {
				checkRes := &auth.CheckCredentialResult{
					TokenValid: false,
				}
				a.
					EXPECT().
					CheckCredential(gomock.Eq(req.Context()), gomock.Eq(checkParam)).
					Return(checkRes, nil).
					Times(1)

				b := &restapp.ResponseBodyInfo{
					Code:    1003,
					Message: "credential is invalid",
				}
				s.
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
					WriteHeader(401).
					Times(1)
				rw.
					EXPECT().
					Write(gomock.Eq([]byte{})).
					Times(1)

				m.ServeHTTP(rw, req)
			})
		})

		When("credential is valid", func() {
			It("should return result", func() {
				a.
					EXPECT().
					CheckCredential(gomock.Eq(req.Context()), gomock.Eq(checkParam)).
					Return(checkRes, nil).
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
