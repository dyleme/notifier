package config

import (
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/telegram/handler"
)

type Config struct {
	Env          string
	DatabaseFile string
	APIKey       string
	NotifierJob  eventnotifier.Config
	Telegram     handler.Config
}

func mapConfig(cc *compositeConfig) Config {
	return Config{
		Env:          cc.Env,
		DatabaseFile: cc.DatabaseFile,
		NotifierJob: eventnotifier.Config{
			CheckTasksPeriod: cc.NotifierJob.CheckPeriod,
		},
		Telegram: handler.Config{
			Token: cc.Telegram.Token,
		},
	}
}
