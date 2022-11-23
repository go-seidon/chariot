package router

import (
	"context"

	"github.com/go-seidon/chariot/internal/storage"
)

type Router interface {
	CreateStorage(ctx context.Context, p CreateStorageParam) (storage.Storage, error)
}

type CreateStorageParam struct {
	BarrelCode string
}

type router struct {
	barrels map[string]storage.Storage
}

func (r *router) CreateStorage(ctx context.Context, p CreateStorageParam) (storage.Storage, error) {
	storage, ok := r.barrels[p.BarrelCode]
	if !ok {
		return nil, ErrUnsupported
	}
	return storage, nil
}

type RouterParam struct {
	Barrels map[string]storage.Storage
}

func NewRouter(p RouterParam) *router {
	return &router{
		barrels: p.Barrels,
	}
}
