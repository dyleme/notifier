package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type collectableConfig struct {
	Database databaseConfig
	JWT      jwtConfig
	APIKey   apiKeyConfig
	Server   serverConfig
}

type serverConfig struct {
	Port                    int           `env:"APP_PORT" env-required:"true"`
	MaxHeaderBytes          int           `env:"MAX_HEADER" env-required:"true"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT" env-required:"true"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT" env-required:"true"`
	TimeForGracefulShutdown time.Duration `env:"GRACEFUL_SHUTDOWN_TIME" env-required:"true"`
}

type databaseConfig struct {
	Port     int    `env:"DB_PORT" env-required:"true"`
	Host     string `env:"DB_HOST" env-required:"true"`
	SSLMode  string `env:"DB_SSL_MODE" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Database string `env:"POSTGRES_DB" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
}

type jwtConfig struct {
	TokenTTL  time.Duration `env:"TOKEN_TTL" env-required:"true"`
	SignedKey string        `env:"SIGNED_KEY" env-required:"true"`
}

type apiKeyConfig struct {
	Key string `env:"API_KEY" env-required:"true"`
}

func Load() Config {
	var collectConfigs collectableConfig
	err := cleanenv.ReadConfig(".env", &collectConfigs)
	if err != nil {
		log.Fatal(err)
	}
	return mapConfig(&collectConfigs)
}
