package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type compositeConfig struct {
	Env         string `env:"ENV" env-required:"true"`
	Database    databaseConfig
	JWT         jwtConfig
	APIKey      apiKeyConfig
	Server      serverConfig
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

type notifierJobConfig struct {
	CheckPeriod time.Duration `env:"TIMETABLE_TASK_CHECK_PERIOD" env-required:"true"`
}

type telegramConfig struct {
	Token string `env:"TELEGRAM_TOKEN" env-required:"true"`
}

func Load() (Config, error) {
	var collectConfigs compositeConfig
	filename := ".env"
	folderPath, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("can't get working dir: %w", err)
	}

	var pathToFile string
	for folderPath != "/" {
		pathToFile = folderPath + "/" + filename
		_, err = os.Stat(pathToFile)
		if err == nil {
			break
		}
		idx := strings.LastIndex(folderPath, "/")
		if idx == -1 {
			break
		}
		folderPath = folderPath[:idx]
	}

	err = cleanenv.ReadConfig(pathToFile, &collectConfigs)
	if err != nil {
		err = cleanenv.ReadEnv(&collectConfigs)
		if err != nil {
			return Config{}, fmt.Errorf("can't read env: %w", err)
		}
	}

	return mapConfig(&collectConfigs), nil
}
