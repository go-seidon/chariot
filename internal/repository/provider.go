package repository

import (
	"context"
)

const (
	PROVIDER_MYSQL = "mysql"
)

type Provider interface {
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	GetAuth() Auth
	GetBarrel() Barrel
}
