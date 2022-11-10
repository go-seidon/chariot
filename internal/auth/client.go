package auth

import (
	"context"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/hashing"
	"github.com/go-seidon/provider/identifier"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/validation"
)

type AuthClient interface {
	CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, *system.SystemError)
	FindClientById(ctx context.Context, p FindClientByIdParam) (*FindClientByIdResult, *system.SystemError)
	UpdateClientById(ctx context.Context, p UpdateClientByIdParam) (*UpdateClientByIdResult, *system.SystemError)
}

type CreateClientParam struct {
	ClientId     string `validate:"required,min=3,max=128" label:"client_id"`
	ClientSecret string `validate:"required,min=8,max=128" label:"client_secret"`
	Name         string `validate:"required,min=3,max=64" label:"name"`
	Type         string `validate:"required,oneof='basic'" label:"type"`
	Status       string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type CreateClientResult struct {
	Success   system.SystemSuccess
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
	Success   system.SystemSuccess
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
	ClientId string `validate:"required,min=3,max=128" label:"client_id"`
	Name     string `validate:"required,min=3,max=64" label:"name"`
	Type     string `validate:"required,oneof='basic'" label:"type"`
	Status   string `validate:"required,oneof='active' 'inactive'" label:"status"`
}

type UpdateClientByIdResult struct {
	Success   system.SystemSuccess
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type authClient struct {
	validator  validation.Validator
	hasher     hashing.Hasher
	identifier identifier.Identifier
	clock      datetime.Clock
	authRepo   repository.Auth
}

func (c *authClient) CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, *system.SystemError) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	id, err := c.identifier.GenerateId()
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	secret, err := c.hasher.Generate(p.ClientSecret)
	if err != nil {
		return nil, &system.SystemError{
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
			return nil, &system.SystemError{
				Code:    status.INVALID_PARAM,
				Message: "client is already exists",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &CreateClientResult{
		Success: system.SystemSuccess{
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

func (c *authClient) FindClientById(ctx context.Context, p FindClientByIdParam) (*FindClientByIdResult, *system.SystemError) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	authClient, err := c.authRepo.FindClient(ctx, repository.FindClientParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.SystemError{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "auth client is not available",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &FindClientByIdResult{
		Success: system.SystemSuccess{
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

func (c *authClient) UpdateClientById(ctx context.Context, p UpdateClientByIdParam) (*UpdateClientByIdResult, *system.SystemError) {
	err := c.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
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
			return nil, &system.SystemError{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "auth client is not available",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &UpdateClientByIdResult{
		Success: system.SystemSuccess{
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

type AuthClientParam struct {
	Validator  validation.Validator
	Hasher     hashing.Hasher
	Identifier identifier.Identifier
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
