package app

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/repository"
	repository_mysql "github.com/go-seidon/chariot/internal/repository-mysql"
	"github.com/go-seidon/provider/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func NewDefaultRepository(config *Config) (repository.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if config.RepositoryProvider != repository.PROVIDER_MYSQL {
		return nil, fmt.Errorf("invalid repository provider")
	}

	var repo repository.Provider
	if config.RepositoryProvider == repository.PROVIDER_MYSQL {
		dbPrimary, err := mysql.NewClient(
			mysql.WithAuth(config.MySQLPrimaryUser, config.MySQLPrimaryPassword),
			mysql.WithConfig(mysql.ClientConfig{DbName: config.MySQLPrimaryDBName}),
			mysql.WithLocation(config.MySQLPrimaryHost, config.MySQLPrimaryPort),
			mysql.ParseTime(),
		)
		if err != nil {
			return nil, err
		}

		dbSecondary, err := mysql.NewClient(
			mysql.WithAuth(config.MySQLSecondaryUser, config.MySQLSecondaryPassword),
			mysql.WithConfig(mysql.ClientConfig{DbName: config.MySQLSecondaryDBName}),
			mysql.WithLocation(config.MySQLSecondaryHost, config.MySQLSecondaryPort),
			mysql.ParseTime(),
		)
		if err != nil {
			return nil, err
		}

		dbClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
			Conn:                      dbPrimary,
			SkipInitializeWithVersion: true,
		}), &gorm.Config{
			DisableAutomaticPing: true,
		})
		if err != nil {
			return nil, err
		}

		err = dbClient.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{
				gorm_mysql.New(gorm_mysql.Config{
					Conn:                      dbSecondary,
					SkipInitializeWithVersion: true,
				}),
			},
		}))
		if err != nil {
			return nil, err
		}

		repo, err = repository_mysql.NewRepository(
			repository_mysql.WithGormClient(dbClient),
		)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}
