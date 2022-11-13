package repository

import (
	"context"
	"time"
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, error)
	FindBarrel(ctx context.Context, p FindBarrelParam) (*FindBarrelResult, error)
}

type CreateBarrelParam struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type CreateBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type FindBarrelParam struct {
	Id string
}

type FindBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
