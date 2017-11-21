package vault

import (
	"github.com/kassisol/libsecret"
	"github.com/kassisol/libsecret/driver"
)

func init() {
	libsecret.RegisterDriver("vault", New)
}

type Config struct {
	Keys map[string]string
}

func New() (driver.Secreter, error) {
	return &Config{
		Keys: make(map[string]string),
	}, nil
}
