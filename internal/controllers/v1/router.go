package v1

import (
	"net/http"

	"CabBookingService/internal/config"
	"CabBookingService/internal/repositories"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func NewV1Router(cfg *config.Config, db *gorm.DB) http.Handler {
	// 1. Init Repositories (Data Layer)
	userRepo := repositories.NewGormUserRepository(db)

	// 2. Init Handlers (Controller Layer)
	userHandler := NewUserHandler(cfg, userRepo)

	// 3. Create the v1 router
	r := chi.NewRouter()

	// --- Define all v1 routes ---

	// Public routes
	r.Post("/register/passenger", userHandler.RegisterPassenger)
	r.Post("/login", userHandler.Login)

	return r
}
