package repository

import (
	"context"
	"time"
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, error)
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
