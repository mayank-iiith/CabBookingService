package v1

import (
	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/domain"
	"CabBookingService/internal/models"
	"CabBookingService/internal/services"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// BookingHandler holds the dependencies for the bookings controllers
type BookingHandler struct {
	bookingService services.BookingService
}

// NewBookingHandler creates a new BookingHandler
func NewBookingHandler(bookingService services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBookingRequest defines the expected JSON body for booking
type CreateBookingRequest struct {
	PickupLatitude   float64    `json:"pickup_latitude"`
	PickupLongitude  float64    `json:"pickup_longitude"`
	DropoffLatitude  float64    `json:"dropoff_latitude"`
	DropoffLongitude float64    `json:"dropoff_longitude"`
	ScheduledTime    *time.Time `json:"scheduled_time"`
}

// CreateBookingResponse defines the JSON response for a successful booking
type CreateBookingResponse struct {
	ID         string               `json:"id"`
	Status     models.BookingStatus `json:"status"`
	PickupLat  float64              `json:"pickup_lat"`
	PickupLon  float64              `json:"pickup_lon"`
	DropoffLat float64              `json:"dropoff_lat"`
	DropoffLon float64              `json:"dropoff_lon"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, ok := r.Context().Value(AccountKey).(*models.Account)
	if !ok {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// TODO: Add Authorization: Only passengers can book

	// 2. Parse Request
	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// TODO: Add Validations (e.g. validate lat/long range)
	//	if req.PickupLat == 0 || req.PickupLon == 0 || req.DropoffLat == 0 || req.DropoffLon == 0 {
	//		helper.RespondWithError(w, http.StatusBadRequest, "All location fields are required")
	//		return
	//	}

	// 3. Construct Params
	params := services.CreateBookingParams{
		PassengerAccountID: account.ID,
		PickupLatitude:     req.PickupLatitude,
		PickupLongitude:    req.PickupLongitude,
		DropoffLatitude:    req.DropoffLatitude,
		DropoffLongitude:   req.DropoffLongitude,
		ScheduledTime:      req.ScheduledTime,
	}

	// 4. Call Service
	booking, err := h.bookingService.CreateBooking(r.Context(), params)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := CreateBookingResponse{
		ID:         booking.ID.String(),
		Status:     booking.Status,
		PickupLat:  booking.PickupLatitude,
		PickupLon:  booking.PickupLongitude,
		DropoffLat: booking.DropoffLatitude,
		DropoffLon: booking.DropoffLongitude,
		CreatedAt:  booking.CreatedAt,
		UpdatedAt:  booking.UpdatedAt,
	}
	helper.RespondWithJSON(w, http.StatusCreated, resp)
}

// RateRequest Struct
type RateRequest struct {
	Rating int    `json:"rating"` // 1-5
	Note   string `json:"note"`
}

// RateRide POST /bookings/{bookingId}/rate
func (h *BookingHandler) RateRide(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, ok := r.Context().Value(AccountKey).(*models.Account)
	if !ok {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Parse bookingId from URL
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// 3. Parse request body
	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		helper.RespondWithError(w, http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}

	// 4. Determine Role (Passenger or Driver?)
	// In a real app, you might check roles explicitly.
	// For now, we assume if they hit this endpoint, they act as Passenger.
	// You should probably create a separate handler for Drivers or check account.Roles

	isPassenger := false
	isDriver := false
	// Check loaded roles from Account
	for _, role := range account.Roles {
		if role.Name == domain.RolePassenger {
			isPassenger = true
		}
		if role.Name == domain.RoleDriver {
			isDriver = true
		}
	}

	// TODO: Logic to handle users who might be BOTH (unlikely in this simple model, but good practice)
	// If the endpoint is hit, we act based on the user's primary intent or role.

	// 5. Call Service to Rate Ride
	// Note: We changed logic slightly. Passing 'isPassenger' boolean tells service
	// "Is the reviewer acting as a passenger?".
	if isPassenger && !isDriver {
		err = h.bookingService.RateRide(r.Context(), bookingID, req.Rating, req.Note, true)
	} else if isDriver && !isPassenger {
		err = h.bookingService.RateRide(r.Context(), bookingID, req.Rating, req.Note, false)
	} else {
		// User has both roles? Usually, you'd check if they were the driver *for this specific booking*
		// But for now, let's default to Passenger if ambiguous, or return error.
		// Better fix: Let the Service decide based on the ID match.

		// Pass the AccountID to the service and let it figure out if this account was the driver or passenger
		// For now, retaining your boolean signature:
		err = h.bookingService.RateRide(r.Context(), bookingID, req.Rating, req.Note, isPassenger)
	}

	helper.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "rating submitted"})
}
