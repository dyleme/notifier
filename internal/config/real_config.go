package config

import (
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
	"github.com/Dyleme/Notifier/internal/server"
)

type Config struct {
	Database *sqldatabase.Config
	JWT      *jwt.Config
	APIKey   string
	Server   *server.Config
}

func mapConfig(collConf *collectableConfig) Config {
	return Config{
		Database: &sqldatabase.Config{
			Port:     collConf.Database.Port,
			Host:     collConf.Database.Host,
			SSLMode:  collConf.Database.SSLMode,
			User:     collConf.Database.User,
			Database: collConf.Database.Database,
			Password: collConf.Database.Password,
		},
		JWT: &jwt.Config{
			SignedKey: collConf.JWT.SignedKey,
			TTL:       collConf.JWT.TokenTTL,
		},
		APIKey: collConf.APIKey.Key,
		Server: &server.Config{
			Port:                    collConf.Server.Port,
			MaxHeaderBytes:          collConf.Server.MaxHeaderBytes,
			ReadTimeout:             collConf.Server.ReadTimeout,
			WriteTimeout:            collConf.Server.WriteTimeout,
			TimeForGracefulShutdown: collConf.Server.TimeForGracefulShutdown,
		},
	}
}
