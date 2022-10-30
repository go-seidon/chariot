package repository_mysql

import (
	"fmt"

	"gorm.io/gorm"
)

type RepositoryParam struct {
	dbClient *gorm.DB
}

type RepoOption = func(*RepositoryParam)

func WithDbClient(c *gorm.DB) RepoOption {
	return func(p *RepositoryParam) {
		p.dbClient = c
	}
}

func NewRepository(opts ...RepoOption) (*provider, error) {
	p := RepositoryParam{}
	for _, opt := range opts {
		opt(&p)
	}

	if p.dbClient == nil {
		return nil, fmt.Errorf("invalid db client specified")
	}

	authRepo := &auth{
		dbClient: p.dbClient,
	}

	repo := &provider{
		dbClient: p.dbClient,
		authRepo: authRepo,
	}
	return repo, nil
}
