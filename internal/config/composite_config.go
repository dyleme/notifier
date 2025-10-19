package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type compositeConfig struct {
	Env          string `env:"ENV"      env-required:"true"`
	DatabaseFile string `env:"DB_FILE"  env-required:"true"`
	LogFile      string `env:"LOG_FILE"`
	NotifierJob  notifierJobConfig
	Telegram     telegramConfig
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
