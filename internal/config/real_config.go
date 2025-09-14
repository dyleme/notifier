package config

import (
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	"github.com/Dyleme/Notifier/internal/httpserver"
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/telegram/handler"
)

type Config struct {
	Env          string
	DatabaseFile string
	JWT          *jwt.Config
	APIKey       string
	Server       *httpserver.Config
	NotifierJob  eventnotifier.Config
	Telegram     handler.Config
}

func mapConfig(cc *compositeConfig) Config {
	return Config{
		Env:          cc.Env,
		DatabaseFile: cc.DatabaseFile,
		JWT: &jwt.Config{
			SignedKey: cc.JWT.SignedKey,
			TTL:       cc.JWT.TokenTTL,
		},
		APIKey: cc.APIKey.Key,
		Server: &httpserver.Config{
			Port:                    cc.Server.Port,
			MaxHeaderBytes:          cc.Server.MaxHeaderBytes,
			ReadTimeout:             cc.Server.ReadTimeout,
			WriteTimeout:            cc.Server.WriteTimeout,
			TimeForGracefulShutdown: cc.Server.TimeForGracefulShutdown,
		},
		NotifierJob: eventnotifier.Config{
			CheckTasksPeriod: cc.NotifierJob.CheckPeriod,
		},
		Telegram: handler.Config{
			Token: cc.Telegram.Token,
		},
	}
}
