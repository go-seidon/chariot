package auth_test

import (
	"context"
	"fmt"
	"time"

	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/repository"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	mock_datetime "github.com/go-seidon/provider/datetime/mock"
	mock_hashing "github.com/go-seidon/provider/hashing/mock"
	mock_identifier "github.com/go-seidon/provider/identifier/mock"
	mock_validation "github.com/go-seidon/provider/validation/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Package", func() {

	Context("CreateClient function", Label("unit"), func() {

		var (
			ctx         context.Context
			currentTs   time.Time
			authClient  auth.AuthClient
			p           auth.CreateClientParam
			validator   *mock_validation.MockValidator
			identifier  *mock_identifier.MockIdentifier
			hasher      *mock_hashing.MockHasher
			clock       *mock_datetime.MockClock
			authRepo    *mock_repository.MockAuth
			createParam repository.CreateClientParam
			createRes   *repository.CreateClientResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			hasher = mock_hashing.NewMockHasher(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			authRepo = mock_repository.NewMockAuth(ctrl)
			authClient = auth.NewAuthClient(auth.AuthClientParam{
				Validator:  validator,
				Hasher:     hasher,
				Identifier: identifier,
				Clock:      clock,
				AuthRepo:   authRepo,
			})
			p = auth.CreateClientParam{
				ClientId:     "client-id",
				ClientSecret: "client-secret",
				Name:         "client-name",
				Type:         "basic",
				Status:       "active",
			}
			createParam = repository.CreateClientParam{
				Id:           "id",
				ClientId:     p.ClientId,
				ClientSecret: "secret",
				Name:         p.Name,
				Type:         p.Type,
				Status:       p.Status,
				CreatedAt:    currentTs,
			}
			createRes = &repository.CreateClientResult{
				Id:           "id",
				ClientId:     p.ClientId,
				ClientSecret: "secret",
				Name:         p.Name,
				Type:         p.Type,
				Status:       p.Status,
				CreatedAt:    currentTs,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed generate id", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("", fmt.Errorf("generate error")).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("generate error"))
			})
		})

		When("failed hash secret", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				hasher.
					EXPECT().
					Generate(gomock.Eq(p.ClientSecret)).
					Return(nil, fmt.Errorf("hash error")).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("hash error"))
			})
		})

		When("failed create client", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				hasher.
					EXPECT().
					Generate(gomock.Eq(p.ClientSecret)).
					Return([]byte("secret"), nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				authRepo.
					EXPECT().
					CreateClient(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("client is already exists", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				hasher.
					EXPECT().
					Generate(gomock.Eq(p.ClientSecret)).
					Return([]byte("secret"), nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				authRepo.
					EXPECT().
					CreateClient(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(nil, repository.ErrExists).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("client is already exists"))
			})
		})

		When("success create client", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				hasher.
					EXPECT().
					Generate(gomock.Eq(p.ClientSecret)).
					Return([]byte("secret"), nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				authRepo.
					EXPECT().
					CreateClient(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				res, err := authClient.CreateClient(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success create auth client"))
				Expect(res.Id).To(Equal("id"))
				Expect(res.Name).To(Equal("client-name"))
				Expect(res.Status).To(Equal("active"))
				Expect(res.Type).To(Equal("basic"))
				Expect(res.CreatedAt).To(Equal(currentTs))
				Expect(err).To(BeNil())
			})
		})
	})

})
