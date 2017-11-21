package stdin

import (
	"github.com/kassisol/docker-volume-git/secret"
	"github.com/kassisol/docker-volume-git/secret/driver"
)

func init() {
	secret.RegisterDriver("stdin", New)
}

type Config struct {
	Keys map[string]string
}

func New() (driver.Secreter, error) {
	return &Config{
		Keys: make(map[string]string),
	}, nil
}
