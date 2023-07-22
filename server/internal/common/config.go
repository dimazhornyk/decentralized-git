package common

import (
	"github.com/caarlos0/env"
	"time"
)

type Config struct {
	Port         int           `env:"PORT" envDefault:"5050"`
	StorageToken string        `env:"STORAGE_TOKEN" envDefault:""`
	JWTSecret    string        `env:"JWT_SECRET" envDefault:""`
	JWTDuration  time.Duration `env:"JWT_DURATION" evnDefault:"24h"`
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	return c, nil
}
