package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"CabBookingService/internal/config"
	"CabBookingService/internal/controllers/v1"
	"CabBookingService/internal/db"
	"CabBookingService/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const (
	v1Route = "/v1"
)

func main() {
	// 1. Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config") // Fatal calls os.Exit(1)
	}
	log.Info().Msg("Config loaded successfully.")

	// 2. Initialize Logger
	logger.Init(logger.Config{
		Environment: cfg.Environment,
		LogLevel:    cfg.LogLevel,
	})

	// 3. Connect to Database
	dbConn, err := db.NewGormDbConn(cfg.PostgresConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Connected to database successfully.")

	// 4. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)                 // Captures panics from handlers to prevent server crash
	r.Use(middleware.Timeout(60 * time.Second)) // Set a timeout for all requests

	// --- Routes ---
	// TODO: Remove this test route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Cab Booking API"))
	})

	// Mount v1 routes
	// All v1.NewV1Router routes will be prefixed with /v1
	r.Mount(v1Route, v1.NewV1Router(cfg, dbConn))

	// --- End Routes ---

	// 5. Server Setup
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.APIPort),
		Handler: r,
	}

	// 6. Start Server - Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Starting Cab Booking API on :%s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	<-quit
	log.Info().Msg("Shutting down server...")

	// Context with timeout for graceful shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(timeoutCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("Server exiting gracefully")

	// TODO: Make sure to close any other resources like DB connections here
}

func init() {
	// --- Load .env file ---
	// This will load the .env file from the root
	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
}
