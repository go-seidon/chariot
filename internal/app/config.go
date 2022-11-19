package app

import (
	"fmt"
	"os"

	"github.com/go-seidon/provider/config/viper"
)

type Config struct {
	AppName    string `env:"APP_NAME"`
	AppEnv     string `env:"APP_ENV"`
	AppVersion string `env:"APP_VERSION"`
	AppDebug   bool   `env:"APP_DEBUG"`

	RESTAppHost string `env:"REST_APP_HOST"`
	RESTAppPort int    `env:"REST_APP_PORT"`

	RepositoryProvider string `env:"REPOSITORY_PROVIDER"`

	MySQLPrimaryHost     string `env:"MYSQL_PRIMARY_HOST"`
	MySQLPrimaryPort     int    `env:"MYSQL_PRIMARY_PORT"`
	MySQLPrimaryUser     string `env:"MYSQL_PRIMARY_USER"`
	MySQLPrimaryPassword string `env:"MYSQL_PRIMARY_PASSWORD"`
	MySQLPrimaryDBName   string `env:"MYSQL_PRIMARY_DB_NAME"`

	MySQLSecondaryHost     string `env:"MYSQL_SECONDARY_HOST"`
	MySQLSecondaryPort     int    `env:"MYSQL_SECONDARY_PORT"`
	MySQLSecondaryUser     string `env:"MYSQL_SECONDARY_USER"`
	MySQLSecondaryPassword string `env:"MYSQL_SECONDARY_PASSWORD"`
	MySQLSecondaryDBName   string `env:"MYSQL_SECONDARY_DB_NAME"`

	UploadFormSize int64 `env:"UPLOAD_FORM_SIZE"`
}

func NewDefaultConfig() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = ENV_LOCAL
	}
	cfg := &Config{AppEnv: appEnv}

	cfgFileName := fmt.Sprintf("config/%s.toml", cfg.AppEnv)
	tomlConfig, err := viper.NewConfig(
		viper.WithFileName(cfgFileName),
	)
	if err != nil {
		return nil, err
	}

	err = tomlConfig.LoadConfig()
	if err != nil {
		return nil, err
	}

	err = tomlConfig.ParseConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
