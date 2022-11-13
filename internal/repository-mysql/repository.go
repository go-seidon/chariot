package repository_mysql

import (
	"fmt"

	db_mysql "github.com/go-seidon/provider/db-mysql"
	"gorm.io/gorm"
)

type RepositoryParam struct {
	gormClient *gorm.DB
	dbClient   db_mysql.Client
}

type RepoOption = func(*RepositoryParam)

func WithGormClient(g *gorm.DB) RepoOption {
	return func(p *RepositoryParam) {
		p.gormClient = g
	}
}

func WithDbClient(c db_mysql.Client) RepoOption {
	return func(p *RepositoryParam) {
		p.dbClient = c
	}
}

func NewRepository(opts ...RepoOption) (*provider, error) {
	p := RepositoryParam{}
	for _, opt := range opts {
		opt(&p)
	}

	if p.dbClient == nil && p.gormClient == nil {
		return nil, fmt.Errorf("invalid db client")
	}

	var err error
	dbClient := p.dbClient
	if dbClient == nil {
		dbClient, err = p.gormClient.DB()
		if err != nil {
			return nil, err
		}
	}

	authRepo := &auth{
		gormClient: p.gormClient,
	}
	barrelRepo := &barrel{
		gormClient: p.gormClient,
	}

	repo := &provider{
		dbClient:   dbClient,
		authRepo:   authRepo,
		barrelRepo: barrelRepo,
	}
	return repo, nil
}
