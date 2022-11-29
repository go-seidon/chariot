package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/go-seidon/chariot/internal/signature"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/typeconv"
	"github.com/golang-jwt/jwt/v4"
)

type jwtSignature struct {
	clock      datetime.Clock
	signMethod jwt.SigningMethod
	signKey    interface{}
	issuer     string
}

func (s *jwtSignature) CreateSignature(ctx context.Context, p signature.CreateSignatureParam) (*signature.CreateSignatureResult, error) {
	var issuedAt time.Time
	if p.IssuedAt != nil {
		issuedAt = typeconv.TimeVal(p.IssuedAt)
	} else {
		issuedAt = s.clock.Now()
	}

	expiresAt := issuedAt.Add(p.Duration)

	claims := struct {
		jwt.RegisteredClaims
		Data map[string]interface{} `json:"data,omitempty"`
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        typeconv.StringVal(p.Id),
		},
		Data: p.Data,
	}

	token := jwt.NewWithClaims(s.signMethod, claims)

	sign, err := token.SignedString(s.signKey)
	if err != nil {
		return nil, err
	}

	res := &signature.CreateSignatureResult{
		Signature: sign,
		IssuedAt:  issuedAt.UTC(),
		ExpiresAt: expiresAt.UTC(),
	}
	return res, nil
}

func (s *jwtSignature) VerifySignature(ctx context.Context, p signature.VerifySignatureParam) (*signature.VerifySignatureResult, error) {
	token, err := jwt.Parse(p.Signature, func(t *jwt.Token) (interface{}, error) {
		if s.signMethod.Alg() != t.Method.Alg() {
			return nil, fmt.Errorf("invalid signing method")
		}
		return s.signKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed parse claims")
	}

	issValid := claims.VerifyIssuer(s.issuer, true)
	if !issValid {
		return nil, fmt.Errorf("issuer is not valid")
	}

	res := &signature.VerifySignatureResult{
		Data: claims,
	}
	return res, nil
}

func NewSignature(opts ...SignatureOption) *jwtSignature {
	p := SignatureParam{
		SignMethod: jwt.SigningMethodHS512,
		Clock:      datetime.NewClock(),
	}
	for _, opt := range opts {
		opt(&p)
	}

	return &jwtSignature{
		clock:      p.Clock,
		signMethod: p.SignMethod,
		signKey:    p.SignKey,
		issuer:     p.Issuer,
	}
}
