package app

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/repository"
	repository_mysql "github.com/go-seidon/chariot/internal/repository-mysql"
	db_mysql "github.com/go-seidon/provider/db-mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// @todo: add unit test
func NewDefaultRepository(config *Config) (repository.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if config.RepositoryProvider != repository.PROVIDER_MYSQL {
		return nil, fmt.Errorf("invalid repository provider")
	}

	var repo repository.Provider
	if config.RepositoryProvider == repository.PROVIDER_MYSQL {
		dbPrimary, err := db_mysql.NewClient(
			db_mysql.WithAuth(config.MySQLPrimaryUser, config.MySQLPrimaryPassword),
			db_mysql.WithConfig(db_mysql.ClientConfig{DbName: config.MySQLPrimaryDBName}),
			db_mysql.WithLocation(config.MySQLPrimaryHost, config.MySQLPrimaryPort),
			db_mysql.ParseTime(),
		)
		if err != nil {
			return nil, err
		}

		dbSecondary, err := db_mysql.NewClient(
			db_mysql.WithAuth(config.MySQLSecondaryUser, config.MySQLSecondaryPassword),
			db_mysql.WithConfig(db_mysql.ClientConfig{DbName: config.MySQLSecondaryDBName}),
			db_mysql.WithLocation(config.MySQLSecondaryHost, config.MySQLSecondaryPort),
			db_mysql.ParseTime(),
		)
		if err != nil {
			return nil, err
		}

		dbClient, err := gorm.Open(mysql.New(mysql.Config{
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
				mysql.New(mysql.Config{
					Conn:                      dbSecondary,
					SkipInitializeWithVersion: true,
				}),
			},
		}))
		if err != nil {
			return nil, err
		}

		repo, err = repository_mysql.NewRepository(
			repository_mysql.WithDbClient(dbClient),
		)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}
