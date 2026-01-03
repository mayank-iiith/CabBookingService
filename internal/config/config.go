package config

import (
	"fmt"

	"CabBookingService/internal/db"

	"github.com/caarlos0/env/v11"
)

type JWTConfig struct {
	JWTSecret    string `env:"JWT_SECRET" envDefault:"default_secret"`
	JWTExpiresIn int64  `env:"JWT_EXPIRES_IN" envDefault:"3600"` // in seconds
}

// Config holds all configuration for the application
type Config struct {
	APIPort string `env:"API_PORT" envDefault:"8080"`

	db.PostgresConfig
	JWTConfig
}

// NewConfig creates a new Config instance by parsing environment variables
func NewConfig() (*Config, error) {
	config := &Config{}
	err := env.Parse(config)
	if err != nil {
		return nil, fmt.Errorf("failed togo parse env vars: %w", err)
	}
	return config, nil
}
