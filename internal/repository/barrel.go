package repository

import (
	"context"
	"time"
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, error)
	FindBarrel(ctx context.Context, p FindBarrelParam) (*FindBarrelResult, error)
	UpdateBarrel(ctx context.Context, p UpdateBarrelParam) (*UpdateBarrelResult, error)
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

type UpdateBarrelParam struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	UpdatedAt time.Time
}

type UpdateBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
