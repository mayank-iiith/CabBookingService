package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds configuration for the logger.
type Config struct {
	Environment string
	LogLevel    string
}

// Init configures the global logger.
func Init(cfg Config) {
	// 1. Set Time Format to Unix (faster/smaller)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// 2. Set Log Level based on cfg.LogLevel
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel // Default to Info if parse fails
	}
	zerolog.SetGlobalLevel(level)

	// 3. Configure Output based on Environment
	if cfg.Environment == "development" {
		// Pretty printing for Development (Human readable)
		//zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON output for Production (Machine readable for systems like Datadog/Splunk)
		// Default output is os.Stderr, which is standard for logs
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	log.Info().
		Str("env", cfg.Environment).
		Str("level", level.String()).
		Msg("Logger initialized")
}
