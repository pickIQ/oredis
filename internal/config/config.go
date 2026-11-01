package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func Load() (Config, error) {
	defaultPort := 8080
	if v := os.Getenv("OREDIS_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid OREDIS_PORT %q: %w", v, err)
		}
		defaultPort = p
	}

	var cfg Config
	flag.IntVar(&cfg.Port, "port", defaultPort, "HTTP server port")
	flag.Parse()
	return cfg, nil
}
