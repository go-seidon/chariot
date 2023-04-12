package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/hashing"
	"github.com/go-seidon/provider/identity"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

type AuthClient interface {
	CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, *system.Error)
	FindClientById(ctx context.Context, p FindClientByIdParam) (*FindClientByIdResult, *system.Error)
	UpdateClientById(ctx context.Context, p UpdateClientByIdParam) (*UpdateClientByIdResult, *system.Error)
	SearchClient(ctx context.Context, p SearchClientParam) (*SearchClientResult, *system.Error)
}

type CreateClientParam struct {
	ClientId     string `validate:"required,lowercase,alphanum,min=6,max=128" label:"client_id"`
	ClientSecret string `validate:"required,printascii,min=8,max=128" label:"client_secret"`
	Name         string `validate:"required,printascii,min=3,max=64" label:"name"`
	Type         string `validate:"required,oneof='basic'" label:"type"`
	Status       string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type CreateClientResult struct {
	Success   system.Success
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
}

type FindClientByIdParam struct {
	Id string `validate:"required,min=5,max=64" label:"id"`
}

type FindClientByIdResult struct {
	Success   system.Success
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type UpdateClientByIdParam struct {
	Id       string `validate:"required,min=5,max=64" label:"id"`
	ClientId string `validate:"required,lowercase,alphanum,min=6,max=128" label:"client_id"`
	Name     string `validate:"required,printascii,min=3,max=64" label:"name"`
	Type     string `validate:"required,oneof='basic'" label:"type"`
	Status   string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type UpdateClientByIdResult struct {
	Success   system.Success
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SearchClientParam struct {
	Keyword    string   `validate:"omitempty,printascii,min=2,max=64" label:"keyword"`
	TotalItems int32    `validate:"numeric,min=1,max=100" label:"total_items"`
	Page       int64    `validate:"numeric,min=1" label:"page"`
	Statuses   []string `validate:"unique,min=0,max=2,dive,oneof='active' 'inactive'" label:"statuses"`
}

type SearchClientResult struct {
	Success system.Success
	Items   []SearchClientItem
	Summary SearchClientSummary
}

type SearchClientItem struct {
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type SearchClientSummary struct {
	TotalItems int64
	Page       int64
}

var _ AuthClient = (*authClient)(nil)

type authClient struct {
	validator  validation.Validator
	hasher     hashing.Hasher
	identifier identity.Identifier
	clock      datetime.Clock
	authRepo   repository.Auth
}

func (c *authClient) CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, *system.Error) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := c.identifier.GenerateId()
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	secret, err := c.hasher.Generate(p.ClientSecret)
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := c.clock.Now()
	createRes, err := c.authRepo.CreateClient(ctx, repository.CreateClientParam{
		Id:           id,
		ClientId:     p.ClientId,
		ClientSecret: string(secret),
		Name:         p.Name,
		Type:         p.Type,
		Status:       p.Status,
		CreatedAt:    currentTs,
	})
	if err != nil {
		if errors.Is(err, repository.ErrExists) {
			return nil, &system.Error{
				Code:    status.INVALID_PARAM,
				Message: "client is already exists",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateClientResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success create auth client",
		},
		Id:        createRes.Id,
		ClientId:  createRes.ClientId,
		Name:      createRes.Name,
		Type:      createRes.Type,
		Status:    createRes.Status,
		CreatedAt: createRes.CreatedAt,
	}
	return res, nil
}

func (c *authClient) FindClientById(ctx context.Context, p FindClientByIdParam) (*FindClientByIdResult, *system.Error) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	authClient, err := c.authRepo.FindClient(ctx, repository.FindClientParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "auth client is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &FindClientByIdResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success find auth client",
		},
		Id:        authClient.Id,
		ClientId:  authClient.ClientId,
		Name:      authClient.Name,
		Type:      authClient.Type,
		Status:    authClient.Status,
		CreatedAt: authClient.CreatedAt,
		UpdatedAt: authClient.UpdatedAt,
	}
	return res, nil
}

func (c *authClient) UpdateClientById(ctx context.Context, p UpdateClientByIdParam) (*UpdateClientByIdResult, *system.Error) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	currentTs := c.clock.Now()
	updateRes, err := c.authRepo.UpdateClient(ctx, repository.UpdateClientParam{
		Id:        p.Id,
		ClientId:  p.ClientId,
		Name:      p.Name,
		Type:      p.Type,
		Status:    p.Status,
		UpdatedAt: currentTs,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "auth client is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &UpdateClientByIdResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success update auth client",
		},
		Id:        updateRes.Id,
		ClientId:  updateRes.ClientId,
		Name:      updateRes.Name,
		Type:      updateRes.Type,
		Status:    updateRes.Status,
		CreatedAt: updateRes.CreatedAt,
		UpdatedAt: updateRes.UpdatedAt,
	}
	return res, nil
}

func (c *authClient) SearchClient(ctx context.Context, p SearchClientParam) (*SearchClientResult, *system.Error) {
	err := c.validator.Validate(p)
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

	searchRes, err := c.authRepo.SearchClient(ctx, repository.SearchClientParam{
		Keyword:  p.Keyword,
		Statuses: p.Statuses,
		Limit:    p.TotalItems,
		Offset:   offset,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	items := []SearchClientItem{}
	for _, authClient := range searchRes.Items {
		items = append(items, SearchClientItem{
			Id:        authClient.Id,
			ClientId:  authClient.ClientId,
			Name:      authClient.Name,
			Type:      authClient.Type,
			Status:    authClient.Status,
			CreatedAt: authClient.CreatedAt,
			UpdatedAt: authClient.UpdatedAt,
		})
	}

	res := &SearchClientResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success search auth client",
		},
		Items: items,
		Summary: SearchClientSummary{
			TotalItems: searchRes.Summary.TotalItems,
			Page:       p.Page,
		},
	}
	return res, nil
}

type AuthClientParam struct {
	Validator  validation.Validator
	Hasher     hashing.Hasher
	Identifier identity.Identifier
	Clock      datetime.Clock
	AuthRepo   repository.Auth
}

func NewAuthClient(p AuthClientParam) *authClient {
	return &authClient{
		validator:  p.Validator,
		hasher:     p.Hasher,
		identifier: p.Identifier,
		clock:      p.Clock,
		authRepo:   p.AuthRepo,
	}
}
