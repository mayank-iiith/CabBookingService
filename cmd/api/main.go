package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "CabBookingService/docs"
	"CabBookingService/internal/config"
	"CabBookingService/internal/controllers/v1"
	"CabBookingService/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	v1Route = "/v1"
)

// @title           Cab Booking Service API
// @version         1.0
// @description     Cab Booking Backend Service in Go.
// @host            localhost:8080
// @BasePath        /v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	// 1. Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	log.Println("Config loaded successfully.")

	// 2. Connect to Database
	dbConn, err := db.NewGormDbConn(cfg.PostgresConfig)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Database connected successfully.")

	// TODO: Remove in production
	dbConn = dbConn.Debug()

	// 3. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second)) // Set a timeout for all requests

	// --- Routes ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Cab Booking API"))
	})

	// Mount v1 routes
	// All v1.NewV1Router routes will be prefixed with /v1
	r.Mount(v1Route, v1.NewV1Router(cfg, dbConn))

	// Swagger UI Route
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// --- End Routes ---

	// 4. Graceful Shutdown Setup
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.APIPort),
		Handler: r,
	}

	// 4. Start Server
	// Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Starting Cab Booking API on :%s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Context with timeout for graceful shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(timeoutCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server gracefully stopped")
}

func init() {
	// --- Load .env file ---
	// This will load the .env file from the root
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
