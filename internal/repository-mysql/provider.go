package repository_mysql

import (
	"context"

	"github.com/go-seidon/chariot/internal/repository"
	db_mysql "github.com/go-seidon/provider/db-mysql"
)

type provider struct {
	dbClient   db_mysql.Pingable
	authRepo   *auth
	barrelRepo *barrel
}

func (p *provider) Init(ctx context.Context) error {
	return nil
}

func (p *provider) Ping(ctx context.Context) error {
	return p.dbClient.PingContext(ctx)
}

func (p *provider) GetAuth() repository.Auth {
	return p.authRepo
}

func (p *provider) GetBarrel() repository.Barrel {
	return p.barrelRepo
}
