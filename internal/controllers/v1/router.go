package v1

import (
	"context"
	"net/http"

	"CabBookingService/internal/config"
	"CabBookingService/internal/domain"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services"
	"CabBookingService/internal/services/queue"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func NewV1Router(cfg *config.Config, db *gorm.DB) http.Handler {
	// 1. Init Repositories (Data Layer)
	roleRepo := repositories.NewGormRoleRepository(db)
	accountRepo := repositories.NewGormAccountRepository(db)
	passengerRepo := repositories.NewGormPassengerRepository(db)
	driverRepo := repositories.NewGormDriverRepository(db)
	bookingRepo := repositories.NewGormBookingRepository(db)
	otpRepo := repositories.NewGormOTPRepository(db)
	reviewRepo := repositories.NewGormReviewRepository(db)
	paymentRepo := repositories.NewGormPaymentRepository(db)

	// 2. Init Core Services
	authService := services.NewAuthService(accountRepo, passengerRepo, driverRepo, roleRepo, db, cfg.JWTSecret, cfg.JWTExpiresIn)
	otpService := services.NewOTPService(otpRepo)
	locationService := services.NewNaiveLocationService(driverRepo)
	paymentService := services.NewPaymentService(paymentRepo)

	// 3. Init Queue
	messageQueue := queue.NewInMemoryQueue()

	// 4. Init Consumers (Workers)
	driverMatchingService := services.NewDriverMatchingService(messageQueue, locationService, bookingRepo, driverRepo)
	err := driverMatchingService.StartConsuming()
	if err != nil {
		// We can use Fatal here because if the consumer fails, the app is broken.
		log.Fatal().Err(err).Msg("Failed to start Driver Matching Consumer")
	}

	schedulingService := services.NewSchedulingService(bookingRepo, messageQueue)
	schedulingService.Start(context.Background())

	// 5. Inject Queue into Booking Service
	bookingService := services.NewBookingService(bookingRepo, driverRepo, passengerRepo, reviewRepo, otpService, locationService, paymentService, messageQueue)

	// 3. Init Handlers (Controller Layer)
	userHandler := NewUserHandler(cfg, authService)
	bookingHandler := NewBookingHandler(bookingService)
	driverHandler := NewDriverHandler(bookingService)
	locationHandler := NewLocationHandler(locationService)

	// 3. Create the v1 router
	r := chi.NewRouter()

	// --- Routes ---

	// Public routes
	r.Post("/register/passenger", userHandler.RegisterPassenger)
	r.Post("/register/driver", userHandler.RegisterDriver)
	r.Post("/login", userHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		// Use the Middleware to populate context with Account details
		r.Use(AuthMiddleware(cfg.JWTSecret, accountRepo))

		// Booking routes
		r.Route("/bookings", func(r chi.Router) {
			r.Use(RequireRoleMiddleware(domain.RolePassenger)) // Only passengers can access these routes

			r.Post("/", bookingHandler.CreateBooking)
			//r.Get("/", bookingHandler.ListMyBookings)
		})

		// Driver routes
		r.Route("/driver/bookings", func(r chi.Router) {
			r.Use(RequireRoleMiddleware(domain.RoleDriver)) // Only drivers can access these routes

			r.Get("/pending", driverHandler.ListPendingRides)
			r.Post("/{bookingId}/accept", driverHandler.AcceptBooking)
			r.Post("/{bookingId}/cancel", driverHandler.CancelBooking)
			r.Post("/{bookingId}/start", driverHandler.StartRide)
			r.Post("/{bookingId}/end", driverHandler.EndRide)
			r.Patch("/availability", driverHandler.ToggleAvailability)
		})

		r.Put("/location/update", locationHandler.UpdateDriverLocation)

	})

	return r
}
