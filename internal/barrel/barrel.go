package barrel

import (
	"context"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identity"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

const (
	STATUS_ACTIVE   = "active"
	STATUS_INACTIVE = "inactive"
)

const (
	PROVIDER_HIPPO    = "goseidon_hippo"
	PROVIDER_AWSS3    = "aws_s3"
	PROVIDER_GCLOUD   = "gcloud_storage"
	PROVIDER_ALICLOUD = "alicloud_oss"
)

var (
	SUPPORTED_PROVIDERS = map[string]bool{
		PROVIDER_HIPPO: true,
	}
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, *system.Error)
	FindBarrelById(ctx context.Context, p FindBarrelByIdParam) (*FindBarrelByIdResult, *system.Error)
	UpdateBarrelById(ctx context.Context, p UpdateBarrelByIdParam) (*UpdateBarrelByIdResult, *system.Error)
	SearchBarrel(ctx context.Context, p SearchBarrelParam) (*SearchBarrelResult, *system.Error)
}

type CreateBarrelParam struct {
	Code     string `validate:"required,lowercase,alphanum,min=6,max=128" label:"code"`
	Name     string `validate:"required,printascii,min=3,max=64" label:"name"`
	Provider string `validate:"required,oneof='goseidon_hippo'" label:"provider"`
	Status   string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type CreateBarrelResult struct {
	Success   system.Success
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type FindBarrelByIdParam struct {
	Id string `validate:"required,min=5,max=64" label:"id"`
}

type FindBarrelByIdResult struct {
	Success   system.Success
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type UpdateBarrelByIdParam struct {
	Id       string `validate:"required,min=5,max=64" label:"id"`
	Code     string `validate:"required,lowercase,alphanum,min=6,max=128" label:"code"`
	Name     string `validate:"required,printascii,min=3,max=64" label:"name"`
	Provider string `validate:"required,oneof='goseidon_hippo'" label:"provider"`
	Status   string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type UpdateBarrelByIdResult struct {
	Success   system.Success
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SearchBarrelParam struct {
	Keyword    string   `validate:"omitempty,printascii,min=2,max=64" label:"keyword"`
	TotalItems int32    `validate:"numeric,min=1,max=100" label:"total_items"`
	Page       int64    `validate:"numeric,min=1" label:"page"`
	Statuses   []string `validate:"unique,min=0,max=2,dive,oneof='active' 'inactive'" label:"statuses"`
	Providers  []string `validate:"unique,min=0,max=4,dive,oneof='goseidon_hippo' 'aws_s3' 'gcloud_storage' 'alicloud_oss'" label:"providers"`
}

type SearchBarrelResult struct {
	Success system.Success
	Items   []SearchBarrelItem
	Summary SearchBarrelSummary
}

type SearchBarrelItem struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type SearchBarrelSummary struct {
	TotalItems int64
	Page       int64
}

type barrel struct {
	validator  validation.Validator
	identifier identity.Identifier
	clock      datetime.Clock
	barrelRepo repository.Barrel
}

func (b *barrel) CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, *system.Error) {
	err := b.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := b.identifier.GenerateId()
	if err != nil {
		return nil, &system.Error{
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
			return nil, &system.Error{
				Code:    status.INVALID_PARAM,
				Message: "barrel is already exists",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateBarrelResult{
		Success: system.Success{
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

func (b *barrel) FindBarrelById(ctx context.Context, p FindBarrelByIdParam) (*FindBarrelByIdResult, *system.Error) {
	err := b.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	authClient, err := b.barrelRepo.FindBarrel(ctx, repository.FindBarrelParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "barrel is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &FindBarrelByIdResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success find barrel",
		},
		Id:        authClient.Id,
		Code:      authClient.Code,
		Name:      authClient.Name,
		Provider:  authClient.Provider,
		Status:    authClient.Status,
		CreatedAt: authClient.CreatedAt,
		UpdatedAt: authClient.UpdatedAt,
	}
	return res, nil
}

func (b *barrel) UpdateBarrelById(ctx context.Context, p UpdateBarrelByIdParam) (*UpdateBarrelByIdResult, *system.Error) {
	err := b.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	currentTs := b.clock.Now()
	updateRes, err := b.barrelRepo.UpdateBarrel(ctx, repository.UpdateBarrelParam{
		Id:        p.Id,
		Code:      p.Code,
		Name:      p.Name,
		Provider:  p.Provider,
		Status:    p.Status,
		UpdatedAt: currentTs,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "barrel is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &UpdateBarrelByIdResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success update barrel",
		},
		Id:        updateRes.Id,
		Code:      updateRes.Code,
		Name:      updateRes.Name,
		Provider:  updateRes.Provider,
		Status:    updateRes.Status,
		CreatedAt: updateRes.CreatedAt,
		UpdatedAt: updateRes.UpdatedAt,
	}
	return res, nil
}

func (b *barrel) SearchBarrel(ctx context.Context, p SearchBarrelParam) (*SearchBarrelResult, *system.Error) {
	err := b.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	offset := int64(0)
	if p.Page > 1 {
		offset = (p.Page - 1) * int64(p.TotalItems)
	}

	searchRes, err := b.barrelRepo.SearchBarrel(ctx, repository.SearchBarrelParam{
		Keyword:   p.Keyword,
		Statuses:  p.Statuses,
		Providers: p.Providers,
		Limit:     p.TotalItems,
		Offset:    offset,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	items := []SearchBarrelItem{}
	for _, barrel := range searchRes.Items {
		items = append(items, SearchBarrelItem{
			Id:        barrel.Id,
			Code:      barrel.Code,
			Name:      barrel.Name,
			Provider:  barrel.Provider,
			Status:    barrel.Status,
			CreatedAt: barrel.CreatedAt,
			UpdatedAt: barrel.UpdatedAt,
		})
	}

	res := &SearchBarrelResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success search barrel",
		},
		Items: items,
		Summary: SearchBarrelSummary{
			TotalItems: searchRes.Summary.TotalItems,
			Page:       p.Page,
		},
	}
	return res, nil
}

type BarrelParam struct {
	Validator  validation.Validator
	Identifier identity.Identifier
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
