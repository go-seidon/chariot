package service

import (
	"context"
	"time"

	"github.com/go-seidon/chariot/internal/signature"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identity"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

type Session interface {
	CreateSession(ctx context.Context, p CreateSessionParam) (*CreateSessionResult, *system.Error)
	VerifySession(ctx context.Context, p VerifySessionParam) (*VerifySessionResult, *system.Error)
}

type CreateSessionParam struct {
	Duration time.Duration `validate:"required,min=1,max=31622400" label:"duration"`
	Features []string      `validate:"required,unique,min=1,max=2,dive,required,oneof='upload_file' 'retrieve_file'" label:"features"`
}

type CreateSessionResult struct {
	Success   system.Success
	CreatedAt time.Time
	ExpiresAt time.Time
	Token     string
}

type VerifySessionParam struct {
	Token   string `validate:"required,min=1" label:"token"`
	Feature string `validate:"required,min=1" label:"feature"`
}

type VerifySessionResult struct {
}

var _ Session = (*sessionService)(nil)

type sessionService struct {
	validator  validation.Validator
	identifier identity.Identifier
	clock      datetime.Clock
	signature  signature.Signature
}

func (s *sessionService) CreateSession(ctx context.Context, p CreateSessionParam) (*CreateSessionResult, *system.Error) {
	err := s.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := s.identifier.GenerateId()
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := s.clock.Now()
	duration := p.Duration * time.Second
	features := map[string]int64{}
	for _, feature := range p.Features {
		features[feature] = currentTs.Add(duration).Unix()
	}

	createSign, err := s.signature.CreateSignature(ctx, signature.CreateSignatureParam{
		Id:       &id,
		IssuedAt: &currentTs,
		Duration: duration,
		Data: map[string]interface{}{
			"features": features,
		},
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateSessionResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success create session",
		},
		CreatedAt: createSign.IssuedAt.UTC(),
		ExpiresAt: createSign.ExpiresAt.UTC(),
		Token:     createSign.Signature,
	}
	return res, nil
}

func (s *sessionService) VerifySession(ctx context.Context, p VerifySessionParam) (*VerifySessionResult, *system.Error) {
	err := s.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	signature, err := s.signature.VerifySignature(ctx, signature.VerifySignatureParam{
		Signature: p.Token,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	data, ok := signature.Data["data"].(map[string]interface{})
	if !ok {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: "invalid signature data",
		}
	}

	features, ok := data["features"].(map[string]interface{})
	if !ok {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: "invalid features data",
		}
	}

	_, ok = features[p.Feature]
	if !ok {
		return nil, &system.Error{
			Code:    status.ACTION_FORBIDDEN,
			Message: "feature is not granted",
		}
	}

	res := &VerifySessionResult{}
	return res, nil
}

type SessionParam struct {
	Validator  validation.Validator
	Identifier identity.Identifier
	Clock      datetime.Clock
	Signature  signature.Signature
}

func NewSession(p SessionParam) *sessionService {
	return &sessionService{
		validator:  p.Validator,
		identifier: p.Identifier,
		clock:      p.Clock,
		signature:  p.Signature,
	}
}
