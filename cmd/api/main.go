package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// --- Load .env file ---
	// This will load the .env file from the root
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// --- Read variables from environment ---
	// os.Getenv() now works because godotenv loaded them
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Default fallback
	}

	// --- Build DSN for the database (example) ---
	// You won't hardcode this string here.
	// Your db connection logic will build it.
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DB")

	// This DSN is what you'll pass to GORM
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort)

	// We can print it just to prove it works
	log.Printf("App DSN: host=%s user=%s ...", dbHost, dbUser)
	log.Printf("GORM DSN: %s", dsn)

	// --- Start Server ---
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Cab Booking API is running!")
	})

	log.Printf("Starting Cab Booking API on :%s", apiPort)
	if err := http.ListenAndServe(":"+apiPort, nil); err != nil {
		panic(err)
	}
}
