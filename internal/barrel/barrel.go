package barrel

import (
	"context"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identifier"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, *system.SystemError)
}

type CreateBarrelParam struct {
	Code     string `validate:"required,lowercase,alphanum,min=6,max=128" label:"code"`
	Name     string `validate:"required,printascii,min=3,max=64" label:"name"`
	Provider string `validate:"required,oneof='goseidon_hippo' 'aws_s3' 'gcloud_storage' 'alicloud_oss'" label:"provider"`
	Status   string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type CreateBarrelResult struct {
	Success   system.SystemSuccess
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type barrel struct {
	validator  validation.Validator
	identifier identifier.Identifier
	clock      datetime.Clock
	barrelRepo repository.Barrel
}

func (b *barrel) CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, *system.SystemError) {
	err := b.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := b.identifier.GenerateId()
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := b.clock.Now()
	createRes, err := b.barrelRepo.CreateBarrel(ctx, repository.CreateBarrelParam{
		Id:        id,
		Code:      p.Code,
		Name:      p.Name,
		Provider:  p.Provider,
		Status:    p.Status,
		CreatedAt: currentTs,
	})
	if err != nil {
		if errors.Is(err, repository.ErrExists) {
			return nil, &system.SystemError{
				Code:    status.INVALID_PARAM,
				Message: "barrel is already exists",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateBarrelResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success create barrel",
		},
		Id:        createRes.Id,
		Code:      createRes.Code,
		Name:      createRes.Name,
		Provider:  createRes.Provider,
		Status:    createRes.Status,
		CreatedAt: createRes.CreatedAt,
	}
	return res, nil
}

type BarrelParam struct {
	Validator  validation.Validator
	Identifier identifier.Identifier
	Clock      datetime.Clock
	BarrelRepo repository.Barrel
}

func NewBarrel(p BarrelParam) *barrel {
	return &barrel{
		validator:  p.Validator,
		identifier: p.Identifier,
		clock:      p.Clock,
		barrelRepo: p.BarrelRepo,
	}
}
