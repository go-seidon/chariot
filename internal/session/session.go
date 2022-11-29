package session

import (
	"context"
	"time"

	"github.com/go-seidon/chariot/internal/signature"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identifier"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

type Session interface {
	CreateSession(ctx context.Context, p CreateSessionParam) (*CreateSessionResult, *system.SystemError)
}

type CreateSessionParam struct {
	Duration time.Duration `validate:"required,min=1" label:"duration"`
	Features []string      `validate:"required,unique,min=1,max=2,dive,required,oneof='upload_file' 'retrieve_file'" label:"features"`
}

type CreateSessionResult struct {
	Success   system.SystemSuccess
	CreatedAt time.Time
	ExpiresAt time.Time
	Token     string
}

type session struct {
	validator  validation.Validator
	identifier identifier.Identifier
	clock      datetime.Clock
	signature  signature.Signature
}

func (s *session) CreateSession(ctx context.Context, p CreateSessionParam) (*CreateSessionResult, *system.SystemError) {
	err := s.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := s.identifier.GenerateId()
	if err != nil {
		return nil, &system.SystemError{
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
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateSessionResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success create session",
		},
		CreatedAt: createSign.IssuedAt.UTC(),
		ExpiresAt: createSign.ExpiresAt.UTC(),
		Token:     createSign.Signature,
	}
	return res, nil
}

type SessionParam struct {
	Validator  validation.Validator
	Identifier identifier.Identifier
	Clock      datetime.Clock
	Signature  signature.Signature
}

func NewSession(p SessionParam) *session {
	return &session{
		validator:  p.Validator,
		identifier: p.Identifier,
		clock:      p.Clock,
		signature:  p.Signature,
	}
}
