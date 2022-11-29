package signature

import (
	"context"
	"time"
)

type Signature interface {
	CreateSignature(ctx context.Context, p CreateSignatureParam) (*CreateSignatureResult, error)
	VerifySignature(ctx context.Context, p VerifySignatureParam) (*VerifySignatureResult, error)
}

type CreateSignatureParam struct {
	Id       *string
	IssuedAt *time.Time
	Duration time.Duration
	Data     map[string]interface{}
}

type CreateSignatureResult struct {
	IssuedAt  time.Time
	ExpiresAt time.Time
	Signature string
}

type VerifySignatureParam struct {
	Signature string
}

type VerifySignatureResult struct {
	Data map[string]interface{}
}
