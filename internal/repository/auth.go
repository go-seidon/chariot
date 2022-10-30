package repository

import (
	"context"
	"time"
)

type Auth interface {
	CreateClient(ctx context.Context, p CreateClientParam) (*CreateClientResult, error)
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
