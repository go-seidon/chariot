package mysql

import (
	"context"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/mysql"
)

type provider struct {
	dbClient   mysql.Pingable
	authRepo   *auth
	barrelRepo *barrel
	filerepo   *file
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

func (p *provider) GetFile() repository.File {
	return p.filerepo
}
