package config

import (
	"fmt"

	"CabBookingService/internal/db"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the application
type Config struct {
	APIPort string `env:"API_PORT" envDefault:"8080"`

	db.PostgresConfig
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
