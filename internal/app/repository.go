package app

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/repository/mysql"
	mysql_client "github.com/go-seidon/provider/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func NewDefaultRepository(config *Config) (repository.Repository, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	if config.RepositoryProvider != repository.PROVIDER_MYSQL {
		return nil, fmt.Errorf("invalid repository provider")
	}

	var repo repository.Repository
	if config.RepositoryProvider == repository.PROVIDER_MYSQL {
		dbPrimary, err := mysql_client.NewClient(
			mysql_client.WithAuth(config.MySQLPrimaryUser, config.MySQLPrimaryPassword),
			mysql_client.WithConfig(mysql_client.ClientConfig{DbName: config.MySQLPrimaryDBName}),
			mysql_client.WithLocation(config.MySQLPrimaryHost, config.MySQLPrimaryPort),
			mysql_client.ParseTime(),
		)
		if err != nil {
			return nil, err
		}

		dbSecondary, err := mysql_client.NewClient(
			mysql_client.WithAuth(config.MySQLSecondaryUser, config.MySQLSecondaryPassword),
			mysql_client.WithConfig(mysql_client.ClientConfig{DbName: config.MySQLSecondaryDBName}),
			mysql_client.WithLocation(config.MySQLSecondaryHost, config.MySQLSecondaryPort),
			mysql_client.ParseTime(),
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

		repo, err = mysql.NewRepository(
			mysql.WithGormClient(dbClient),
		)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}
