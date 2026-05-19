package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Cleaner Cleaner `envPrefix:"CLEANER_"`
}

type Cleaner struct {
	Interval time.Duration `env:"INTERVAL" envDefault:"4h"`
}

func Load() (Config, error) {
	return env.ParseAsWithOptions[Config](env.Options{
		Prefix: "NODE_AGENT_",
	})
}
