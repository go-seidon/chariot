package repository

import (
	"context"
	"time"
)

type Auth interface {
	CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, error)
	FindClient(ctx context.Context, p FindClientParam) (*FindClientResult, error)
	UpdateClient(ctx context.Context, p UpdateClientParam) (*UpdateClientResult, error)
	SearchClient(ctx context.Context, p SearchClientParam) (*SearchClientResult, error)
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

type SearchClientParam struct {
	Limit    int32
	Offset   int64
	Keyword  string
	Statuses []string
}

type SearchClientResult struct {
	Summary SearchClientSummary
	Items   []SearchClientItem
}

type SearchClientSummary struct {
	TotalItems int64
}

type SearchClientItem struct {
	Id           string
	ClientId     string
	ClientSecret string
	Name         string
	Type         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
