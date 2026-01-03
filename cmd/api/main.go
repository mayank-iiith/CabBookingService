package main

import (
	"fmt"
	"log"
	"net/http"

	"CabBookingService/internal/config"
	"CabBookingService/internal/controllers/v1"
	"CabBookingService/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

const (
	v1Route = "/v1"
)

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

	// 3. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- Routes ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Cab Booking API"))
	})

	// Mount v1 routes
	// All v1.NewV1Router routes will be prefixed with /v1
	r.Mount(v1Route, v1.NewV1Router(cfg, dbConn))

	// --- End Routes ---

	// 4. Start Server
	log.Printf("Starting Cab Booking API on :%s", cfg.APIPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.APIPort), r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func init() {
	// --- Load .env file ---
	// This will load the .env file from the root
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
