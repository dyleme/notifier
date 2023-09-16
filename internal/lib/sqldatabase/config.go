package sqldatabase

import (
	"fmt"
	"net"
	"strconv"
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
	address := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	return fmt.Sprintf("postgres://%s:%s@%s/%s", c.User, c.Password, address, c.Database)
}
