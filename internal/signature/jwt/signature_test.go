package jwt_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-seidon/chariot/internal/signature"
	"github.com/go-seidon/chariot/internal/signature/jwt"
	datetime "github.com/go-seidon/provider/datetime/mock"
	"github.com/go-seidon/provider/typeconv"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJWT(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JWT Package")
}

var _ = Describe("Signature Package", func() {

	Context("CreateSignature function", Label("unit"), func() {
		var (
			signer    signature.Signature
			ctx       context.Context
			currentTs time.Time
			p         signature.CreateSignatureParam
			clock     *datetime.MockClock
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			clock = datetime.NewMockClock(ctrl)
			signer = jwt.NewSignature(
				jwt.WithIssuer("issuer"),
				jwt.WithSignKey([]byte("key")),
				jwt.WithClock(clock),
			)
			ctx = context.Background()
			currentTs = time.Now().UTC()
			p = signature.CreateSignatureParam{
				Id:       typeconv.String("id"),
				IssuedAt: typeconv.Time(currentTs),
				Duration: 10 * time.Second,
				Data: map[string]interface{}{
					"features": map[string]int64{
						"upload_file": 123,
					},
				},
			}
		})

		When("failed sign", func() {
			It("should return error", func() {
				signer := jwt.NewSignature()
				res, err := signer.CreateSignature(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("key is of invalid type")))
				Expect(res).To(BeNil())
			})
		})

		When("success sign", func() {
			It("should return result", func() {
				res, err := signer.CreateSignature(ctx, p)

				expiresTs := currentTs.Add(p.Duration).UTC()
				Expect(err).To(BeNil())
				Expect(res.Signature).ToNot(Equal(""))
				Expect(res.IssuedAt).To(Equal(currentTs))
				Expect(res.ExpiresAt).To(Equal(expiresTs))
			})
		})

		When("success sign with custom issued at", func() {
			It("should return result", func() {
				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				p := signature.CreateSignatureParam{
					Id:       typeconv.String("id"),
					IssuedAt: nil,
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				}
				res, err := signer.CreateSignature(ctx, p)

				expiresTs := currentTs.Add(p.Duration).UTC()
				Expect(err).To(BeNil())
				Expect(res.Signature).ToNot(Equal(""))
				Expect(res.IssuedAt).To(Equal(currentTs))
				Expect(res.ExpiresAt).To(Equal(expiresTs))
			})
		})
	})

	Context("VerifySignature function", Label("unit"), func() {
		var (
			signer    signature.Signature
			ctx       context.Context
			currentTs time.Time
			// p      signature.VerifySignatureParam
			clock *datetime.MockClock
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			clock = datetime.NewMockClock(ctrl)
			signer = jwt.NewSignature(
				jwt.WithIssuer("issuer"),
				jwt.WithSignKey([]byte("key")),
				jwt.WithClock(clock),
			)
			ctx = context.Background()
			currentTs = time.Now().UTC()
			// p = signature.VerifySignatureParam{
			// 	Signature: "invalid-sign",
			// }
		})

		When("token is invalid", func() {
			It("should return error", func() {
				p := signature.VerifySignatureParam{
					Signature: "invalid-sign",
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err.Error()).To(Equal("token contains an invalid number of segments"))
				Expect(res).To(BeNil())
			})
		})

		When("sign method is invalid", func() {
			It("should return error", func() {
				othSigner := jwt.NewSignature(
					jwt.WithIssuer("issuer"),
					jwt.WithSignKey([]byte("key")),
					jwt.WithSignMethod(gojwt.SigningMethodHS256),
				)
				createSign, _ := othSigner.CreateSignature(ctx, signature.CreateSignatureParam{
					IssuedAt: typeconv.Time(currentTs),
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				})

				p := signature.VerifySignatureParam{
					Signature: createSign.Signature,
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err.Error()).To(Equal("invalid signing method"))
				Expect(res).To(BeNil())
			})
		})

		When("issuer is invalid", func() {
			It("should return error", func() {
				othSigner := jwt.NewSignature(
					jwt.WithIssuer("invalid-issuer"),
					jwt.WithSignKey([]byte("key")),
				)
				createSign, _ := othSigner.CreateSignature(ctx, signature.CreateSignatureParam{
					IssuedAt: typeconv.Time(currentTs),
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				})

				p := signature.VerifySignatureParam{
					Signature: createSign.Signature,
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("issuer is not valid")))
				Expect(res).To(BeNil())
			})
		})

		When("token is not valid yet", func() {
			It("should return error", func() {
				createSign, _ := signer.CreateSignature(ctx, signature.CreateSignatureParam{
					IssuedAt: typeconv.Time(currentTs.Add(20 * time.Minute)),
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				})

				p := signature.VerifySignatureParam{
					Signature: createSign.Signature,
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err.Error()).To(Equal("Token is not valid yet"))
				Expect(res).To(BeNil())
			})
		})

		When("token is expired", func() {
			It("should return error", func() {
				createSign, _ := signer.CreateSignature(ctx, signature.CreateSignatureParam{
					IssuedAt: typeconv.Time(currentTs.Add(-20 * time.Minute)),
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				})

				p := signature.VerifySignatureParam{
					Signature: createSign.Signature,
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err.Error()).To(Equal("Token is expired"))
				Expect(res).To(BeNil())
			})
		})

		When("success verify signature", func() {
			It("should return error", func() {
				createSign, _ := signer.CreateSignature(ctx, signature.CreateSignatureParam{
					IssuedAt: typeconv.Time(currentTs),
					Duration: 10 * time.Second,
					Data: map[string]interface{}{
						"features": map[string]int64{
							"upload_file": 123,
						},
					},
				})

				p := signature.VerifySignatureParam{
					Signature: createSign.Signature,
				}
				res, err := signer.VerifySignature(ctx, p)

				Expect(err).To(BeNil())
				Expect(res.Data["data"]).To(Equal(map[string]interface{}{
					"features": map[string]interface{}{
						"upload_file": float64(123),
					},
				}))
			})
		})
	})
})
