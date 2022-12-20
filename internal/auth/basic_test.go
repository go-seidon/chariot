package auth_test

import (
	"context"
	"fmt"

	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/repository"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	mock_encoding "github.com/go-seidon/provider/encoding/mock"
	mock_hashing "github.com/go-seidon/provider/hashing/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth Package", func() {
	Context("ParseAuthToken function", Label("unit"), func() {
		var (
			ctx       context.Context
			authRepo  *mock_repository.MockAuth
			encoder   *mock_encoding.MockEncoder
			hasher    *mock_hashing.MockHasher
			basicAuth auth.BasicAuth
			p         auth.ParseAuthTokenParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			authRepo = mock_repository.NewMockAuth(ctrl)
			encoder = mock_encoding.NewMockEncoder(ctrl)
			hasher = mock_hashing.NewMockHasher(ctrl)
			basicAuth = auth.NewBasicAuth(auth.NewBasicAuthParam{
				AuthRepo: authRepo,
				Encoder:  encoder,
				Hasher:   hasher,
			})
			p = auth.ParseAuthTokenParam{
				Token: "mock-token",
			}
		})

		When("token is empty", func() {
			It("should return error", func() {
				p.Token = ""
				res, err := basicAuth.ParseAuthToken(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid token")))
			})
		})

		When("failed decode token", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.Token)).
					Return(nil, fmt.Errorf("error decode")).
					Times(1)

				res, err := basicAuth.ParseAuthToken(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error decode")))
			})
		})

		When("encoding is invalid", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.Token)).
					Return([]byte(""), nil).
					Times(1)

				res, err := basicAuth.ParseAuthToken(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid auth encoding")))
			})
		})

		When("client id is invalid", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.Token)).
					Return([]byte(" : "), nil).
					Times(1)

				res, err := basicAuth.ParseAuthToken(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid client id")))
			})
		})

		When("client secret is invalid", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.Token)).
					Return([]byte("client_id: "), nil).
					Times(1)

				res, err := basicAuth.ParseAuthToken(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid client secret")))
			})
		})

		When("success parse auth token", func() {
			It("should return result", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.Token)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				res, err := basicAuth.ParseAuthToken(ctx, p)

				expectedRes := &auth.ParseAuthTokenResult{
					ClientId:     "client_id",
					ClientSecret: "client_secret",
				}
				Expect(res).To(Equal(expectedRes))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("CheckCredential function", Label("unit"), func() {
		var (
			ctx       context.Context
			authRepo  *mock_repository.MockAuth
			encoder   *mock_encoding.MockEncoder
			hasher    *mock_hashing.MockHasher
			basicAuth auth.BasicAuth
			p         auth.CheckCredentialParam
			findParam repository.FindClientParam
			findRes   *repository.FindClientResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			authRepo = mock_repository.NewMockAuth(ctrl)
			encoder = mock_encoding.NewMockEncoder(ctrl)
			hasher = mock_hashing.NewMockHasher(ctrl)
			basicAuth = auth.NewBasicAuth(auth.NewBasicAuthParam{
				AuthRepo: authRepo,
				Encoder:  encoder,
				Hasher:   hasher,
			})
			p = auth.CheckCredentialParam{
				AuthToken: "mock-token",
			}
			findParam = repository.FindClientParam{
				ClientId: "client_id",
			}
			findRes = &repository.FindClientResult{
				Status:       "active",
				ClientId:     "client_id",
				ClientSecret: "hashed_client_secret",
			}
		})

		When("failed parse token", func() {
			It("should return error", func() {
				p.AuthToken = ""
				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid token")))
			})
		})

		When("failed find client", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.AuthToken)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				authRepo.
					EXPECT().
					FindClient(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, fmt.Errorf("db error")).
					Times(1)

				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("db error")))
			})
		})

		When("client is not found", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.AuthToken)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				authRepo.
					EXPECT().
					FindClient(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res.IsValid()).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})

		When("client is inactive", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.AuthToken)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				findRes := &repository.FindClientResult{
					ClientId:     "client_id",
					ClientSecret: "hashed_client_secret",
					Status:       "inactive",
				}
				authRepo.
					EXPECT().
					FindClient(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res.IsValid()).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})

		When("client secret is invalid", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.AuthToken)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				authRepo.
					EXPECT().
					FindClient(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				hasher.
					EXPECT().
					Verify(gomock.Eq(findRes.ClientSecret), gomock.Eq("client_secret")).
					Return(fmt.Errorf("invalid")).
					Times(1)

				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res.IsValid()).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})

		When("client secret is valid", func() {
			It("should return result", func() {
				encoder.
					EXPECT().
					Decode(gomock.Eq(p.AuthToken)).
					Return([]byte("client_id:client_secret"), nil).
					Times(1)

				authRepo.
					EXPECT().
					FindClient(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				hasher.
					EXPECT().
					Verify(gomock.Eq(findRes.ClientSecret), gomock.Eq("client_secret")).
					Return(nil).
					Times(1)

				res, err := basicAuth.CheckCredential(ctx, p)

				Expect(res.IsValid()).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	})
})
