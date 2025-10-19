package config

import (
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/telegram"
)

type Config struct {
	Env          string
	LogFile      string
	DatabaseFile string
	NotifierJob  eventnotifier.Config
	Telegram     telegram.Config
}

func mapConfig(cc *compositeConfig) Config {
	return Config{
		Env:          cc.Env,
		LogFile:      cc.LogFile,
		DatabaseFile: cc.DatabaseFile,
		NotifierJob: eventnotifier.Config{
			CheckTasksPeriod: cc.NotifierJob.CheckPeriod,
		},
		Telegram: telegram.Config{
			Token: cc.Telegram.Token,
		},
	}
}
