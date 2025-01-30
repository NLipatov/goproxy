package config

import (
	"os"
)

type Config struct {
	Mode string
}

func Load() (Config, error) {
	return Config{
		Mode: os.Getenv("MODE"),
	}, nil
}
