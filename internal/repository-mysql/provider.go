package repository_mysql

import (
	"context"

	"github.com/go-seidon/chariot/internal/repository"
	"gorm.io/gorm"
)

type provider struct {
	dbClient *gorm.DB
	authRepo *auth
}

func (p *provider) Init(ctx context.Context) error {
	return nil
}

func (p *provider) Ping(ctx context.Context) error {
	db, err := p.dbClient.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func (p *provider) GetAuth() repository.Auth {
	return p.authRepo
}
