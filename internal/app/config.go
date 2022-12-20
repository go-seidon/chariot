package app

import (
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

	StorageAccessHost string `env:"STORAGE_ACCESS_HOST"`
	StorageFormSize   int64  `env:"STORAGE_FORM_SIZE"`

	QueueProvider         string `env:"QUEUE_PROVIDER"`
	QueueRabbitmqProto    string `env:"QUEUE_RABBITMQ_PROTO"`
	QueueRabbitmqHost     string `env:"QUEUE_RABBITMQ_HOST"`
	QueueRabbitmqPort     int    `env:"QUEUE_RABBITMQ_PORT"`
	QueueRabbitmqUser     string `env:"QUEUE_RABBITMQ_USER"`
	QueueRabbitmqPassword string `env:"QUEUE_RABBITMQ_PASSWORD"`

	SignatureIssuer string `env:"SIGNATURE_ISSUER"`
	SignatureKey    string `env:"SIGNATURE_KEY"`

	Barrels map[string]BarrelConfig `env:"BARRELS"`
}

type BarrelConfig struct {
	Provider          string `env:"PROVIDER"`
	HippoServerHost   string `env:"HIPPO_SERVER_HOST"`
	HippoClientId     string `env:"HIPPO_CLIENT_ID"`
	HippoClientSecret string `env:"HIPPO_CLIENT_SECRET"`
}

func NewDefaultConfig() (*Config, error) {
	tomlConfig, err := viper.NewConfig(
		viper.WithFileName("config/default.toml"),
	)
	if err != nil {
		return nil, err
	}

	err = tomlConfig.LoadConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = tomlConfig.ParseConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
