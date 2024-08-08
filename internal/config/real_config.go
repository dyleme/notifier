package config

import (
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/server"
	"github.com/Dyleme/Notifier/internal/telegram/handler"
	"github.com/Dyleme/Notifier/pkg/sqldatabase"
)

type Config struct {
	Env         string
	Database    *sqldatabase.Config
	JWT         *jwt.Config
	APIKey      string
	Server      *server.Config
	NotifierJob eventnotifier.Config
	Telegram    handler.Config
}

func mapConfig(cc *compositeConfig) Config {
	return Config{
		Env: cc.Env,
		Database: &sqldatabase.Config{
			Port:     cc.Database.Port,
			Host:     cc.Database.Host,
			SSLMode:  cc.Database.SSLMode,
			User:     cc.Database.User,
			Database: cc.Database.Database,
			Password: cc.Database.Password,
		},
		JWT: &jwt.Config{
			SignedKey: cc.JWT.SignedKey,
			TTL:       cc.JWT.TokenTTL,
		},
		APIKey: cc.APIKey.Key,
		Server: &server.Config{
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
