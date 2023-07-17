package sqldatabase

import (
	"fmt"
)

type Config struct {
	Port     int
	Host     string
	SSLMode  string
	User     string
	Database string
	Password string
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Database)
}
