package repository

import (
	"context"
	"time"
)

type Auth interface {
	CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, error)
	FindClient(ctx context.Context, p FindClientParam) (*FindClientResult, error)
	UpdateClient(ctx context.Context, p UpdateClientParam) (*UpdateClientResult, error)
}

type CreateClientParam struct {
	Id           string
	ClientId     string
	ClientSecret string
	Name         string
	Type         string
	Status       string
	CreatedAt    time.Time
}

type CreateClientResult struct {
	Id           string
	ClientId     string
	ClientSecret string
	Name         string
	Type         string
	Status       string
	CreatedAt    time.Time
}

type FindClientParam struct {
	Id string
}

type FindClientResult struct {
	Id           string
	ClientId     string
	ClientSecret string
	Name         string
	Type         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type UpdateClientParam struct {
	Id        string
	ClientId  string
	Name      string
	Type      string
	Status    string
	UpdatedAt time.Time
}

type UpdateClientResult struct {
	Id           string
	ClientId     string
	ClientSecret string
	Name         string
	Type         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
