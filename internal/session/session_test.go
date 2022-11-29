package session_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/chariot/internal/signature"
	mock_signature "github.com/go-seidon/chariot/internal/signature/mock"
	mock_datetime "github.com/go-seidon/provider/datetime/mock"
	mock_identifier "github.com/go-seidon/provider/identifier/mock"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/typeconv"
	mock_validation "github.com/go-seidon/provider/validation/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSession(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Session Package")
}

var _ = Describe("Session Package", func() {

	Context("CreateSession function", Label("unit"), func() {
		var (
			ctx           context.Context
			currentTs     time.Time
			sessionClient session.Session
			p             session.CreateSessionParam
			r             *session.CreateSessionResult
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			signer        *mock_signature.MockSignature
			createParam   signature.CreateSignatureParam
			createRes     *signature.CreateSignatureResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			signer = mock_signature.NewMockSignature(ctrl)
			sessionClient = session.NewSession(session.SessionParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Signature:  signer,
			})

			expiresAt := currentTs.Add(p.Duration * time.Second)
			createParam = signature.CreateSignatureParam{
				Id:       typeconv.String("id"),
				IssuedAt: typeconv.Time(currentTs),
				Duration: p.Duration * time.Second,
				Data: map[string]interface{}{
					"features": map[string]int64{
						"upload_file":   expiresAt.Unix(),
						"retrieve_file": expiresAt.Unix(),
					},
				},
			}
			createRes = &signature.CreateSignatureResult{
				IssuedAt:  currentTs.UTC(),
				ExpiresAt: expiresAt.UTC(),
				Signature: "signature",
			}
			p = session.CreateSessionParam{
				Duration: 600,
				Features: []string{"upload_file", "retrieve_file"},
			}
			r = &session.CreateSessionResult{
				Success: system.SystemSuccess{
					Code:    1000,
					Message: "success create session",
				},
				CreatedAt: currentTs.UTC(),
				ExpiresAt: expiresAt.UTC(),
				Token:     "signature",
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := sessionClient.CreateSession(ctx, p)

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
					Return("", fmt.Errorf("disk error")).
					Times(1)

				res, err := sessionClient.CreateSession(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("disk error"))
			})
		})

		When("failed create signature", func() {
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

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				signer.
					EXPECT().
					CreateSignature(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(nil, fmt.Errorf("config error")).
					Times(1)

				res, err := sessionClient.CreateSession(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("config error"))
			})
		})

		When("success create signature", func() {
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

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				signer.
					EXPECT().
					CreateSignature(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				res, err := sessionClient.CreateSession(ctx, p)

				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})
	})
})
