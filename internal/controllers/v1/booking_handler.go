package v1

import (
	"CabBookingService/internal/controllers/helper"
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
	PickupLat  float64 `json:"pickup_lat"`
	PickupLon  float64 `json:"pickup_lon"`
	DropoffLat float64 `json:"dropoff_lat"`
	DropoffLon float64 `json:"dropoff_lon"`
}

// CreateBookingResponse defines the JSON response for a successful booking
type CreateBookingResponse struct {
	ID           string               `json:"id"`
	Status       models.BookingStatus `json:"status"`
	PickupLat    float64              `json:"pickup_lat"`
	PickupLon    float64              `json:"pickup_lon"`
	DropoffLat   float64              `json:"dropoff_lat"`
	DropoffLon   float64              `json:"dropoff_lon"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	RideStartOTP *string              `json:"ride_start_otp,omitempty"`
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, ok := r.Context().Value(AccountKey).(*models.Account)
	if !ok {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// TODO: Add Validations

	booking, err := h.bookingService.CreateBooking(r.Context(), account.ID, req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := CreateBookingResponse{
		ID:           booking.ID.String(),
		Status:       booking.Status,
		PickupLat:    booking.PickupLatitude,
		PickupLon:    booking.PickupLongitude,
		DropoffLat:   booking.DropoffLatitude,
		DropoffLon:   booking.DropoffLongitude,
		CreatedAt:    booking.CreatedAt,
		UpdatedAt:    booking.UpdatedAt,
		RideStartOTP: &booking.RideStartOTP.Code,
	}
	helper.RespondWithJSON(w, http.StatusCreated, resp)
}

// RateRequest Struct
type RateRequest struct {
	Rating int    `json:"rating"` // 1-5
	Note   string `json:"note"`
}

// RateRide godoc
// @Summary      Rate a completed ride
// @Tags         Passenger
// @Security     BearerAuth
// @Param        bookingId path string true "Booking ID"
// @Param        request body RateRequest true "Rating"
// @Success      200  {object}  map[string]string
// @Router       /bookings/{bookingId}/rate [post]
func (h *BookingHandler) RateRide(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	//account, ok := r.Context().Value(AccountKey).(*models.Account)
	//if !ok {
	//	helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
	//	return
	//}

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

	isPassenger := true // Implementation goes here

	// 5. Call Service to Rate Ride
	if isPassenger {
		err = h.bookingService.RateRide(r.Context(), bookingID, req.Rating, req.Note, isPassenger)
	} else {
		err = h.bookingService.RateRide(r.Context(), bookingID, req.Rating, req.Note, !isPassenger)
	}

	helper.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "rating submitted"})
}

//func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
//	// 1. Get the authenticated user from context (middleware should have set this)
//	user, ok := r.Context().Value(UserKey).(*models.User)
//	if !ok {
//		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
//		return
//	}
//
//	// 2. Authorization: Only passengers can book
//	if !user.IsPassenger {
//		helper.RespondWithError(w, http.StatusForbidden, "Only passengers can request rides")
//		return
//	}
//
//	// 3. Parse and validate the request
//	var req CreateBookingRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
//		return
//	}
//
//	// 4. Validate input
//	if req.PickupLat == 0 || req.PickupLon == 0 || req.DropoffLat == 0 || req.DropoffLon == 0 {
//		helper.RespondWithError(w, http.StatusBadRequest, "All location fields are required")
//		return
//	}
//	// TODO: Add more validation (e.g., valid lat/lon ranges)
//
//	// 5. Create the booking model
//	now := time.Now().UTC()
//	booking := &models.Booking{
//		ID:               uuid.New(),
//		PassengerID:      user.ID,
//		Status:           models.BookingStatusRequested,
//		PickupLatitude:   req.PickupLat,
//		PickupLongitude:  req.PickupLon,
//		DropoffLatitude:  req.DropoffLat,
//		DropoffLongitude: req.DropoffLon,
//		CreatedAt:        now,
//		UpdatedAt:        now,
//	}
//
//	// 6. Save the booking using the repository
//	if err := h.bookingRepo.Create(booking); err != nil {
//		helper.RespondWithError(w, http.StatusInternalServerError, "Failed to create booking")
//		return
//	}
//
//	// 7. Prepare and send the response
//	resp := CreateBookingResponse{
//		ID:         booking.ID.String(),
//		Status:     booking.Status,
//		PickupLat:  booking.PickupLatitude,
//		PickupLon:  booking.PickupLongitude,
//		DropoffLat: booking.DropoffLatitude,
//		DropoffLon: booking.DropoffLongitude,
//		CreatedAt:  booking.CreatedAt,
//		UpdatedAt:  booking.UpdatedAt,
//	}
//	helper.RespondWithJSON(w, http.StatusCreated, resp)
//}
