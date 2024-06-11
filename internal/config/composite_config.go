package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type compositeConfig struct {
	Env         string `env:"ENV" env-required:"true"`
	Database    databaseConfig
	JWT         jwtConfig
	APIKey      apiKeyConfig
	Server      serverConfig
	Notifier    eventConfig
	NotifierJob notifierJobConfig
	Telegram    telegramConfig
}

type serverConfig struct {
	Port                    int           `env:"APP_PORT"               env-required:"true"`
	MaxHeaderBytes          int           `env:"MAX_HEADER"             env-required:"true"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT"           env-required:"true"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT"          env-required:"true"`
	TimeForGracefulShutdown time.Duration `env:"GRACEFUL_SHUTDOWN_TIME" env-required:"true"`
}

type databaseConfig struct {
	Port     int    `env:"DB_PORT"           env-required:"true"`
	Host     string `env:"DB_HOST"           env-required:"true"`
	SSLMode  string `env:"DB_SSL_MODE"       env-required:"true"`
	User     string `env:"POSTGRES_USER"     env-required:"true"`
	Database string `env:"POSTGRES_DB"       env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
}

type jwtConfig struct {
	TokenTTL  time.Duration `env:"TOKEN_TTL"  env-required:"true"`
	SignedKey string        `env:"SIGNED_KEY" env-required:"true"`
}

type apiKeyConfig struct {
	Key string `env:"API_KEY" env-required:"true"`
}

type eventConfig struct {
	CheckPeriod time.Duration `env:"NOTIFICATIONS_CHECK_PERIOD" env-required:"true"`
}

type notifierJobConfig struct {
	CheckPeriod time.Duration `env:"TIMETABLE_TASK_CHECK_PERIOD" env-required:"true"`
}

type telegramConfig struct {
	Token string `env:"TELEGRAM_TOKEN" env-required:"true"`
}

func Load() Config {
	var collectConfigs compositeConfig
	err := cleanenv.ReadConfig(".env", &collectConfigs)
	if err != nil {
		err = cleanenv.ReadEnv(&collectConfigs)
		if err != nil {
			panic(fmt.Errorf("can't read env: %w", err))
		}
	}

	return mapConfig(&collectConfigs)
}
