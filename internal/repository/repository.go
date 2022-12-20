package repository

import (
	"context"
)

const (
	PROVIDER_MYSQL = "mysql"
)

type Repository interface {
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	GetAuth() Auth
	GetBarrel() Barrel
	GetFile() File
}
