package jwt

import (
	"github.com/go-seidon/provider/datetime"
	"github.com/golang-jwt/jwt/v4"
)

type SignatureParam struct {
	Clock      datetime.Clock
	SignMethod jwt.SigningMethod
	SignKey    interface{}
	Issuer     string
}

type SignatureOption = func(*SignatureParam)

func WithClock(c datetime.Clock) SignatureOption {
	return func(p *SignatureParam) {
		p.Clock = c
	}
}

func WithSignMethod(sm jwt.SigningMethod) SignatureOption {
	return func(p *SignatureParam) {
		p.SignMethod = sm
	}
}

func WithSignKey(sk interface{}) SignatureOption {
	return func(p *SignatureParam) {
		p.SignKey = sk
	}
}

func WithIssuer(iss string) SignatureOption {
	return func(p *SignatureParam) {
		p.Issuer = iss
	}
}
