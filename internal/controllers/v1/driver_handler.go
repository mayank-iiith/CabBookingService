package v1

import (
	"encoding/json"
	"net/http"

	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/models"
	"CabBookingService/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DriverHandler struct {
	bookingService services.BookingService
}

// NewDriverHandler creates a new DriverHandler
func NewDriverHandler(bookingService services.BookingService) *DriverHandler {
	return &DriverHandler{
		bookingService: bookingService,
	}
}

// --- Response Structs ---

type DriverActionResponse struct {
	BookingID string `json:"booking_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// AcceptBooking - POST /v1/driver/bookings/{bookingId}/accept
func (h *DriverHandler) AcceptBooking(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Get Booking ID from URL params
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// 3. Call Service to accept booking
	err = h.bookingService.AcceptBooking(r.Context(), account.ID, bookingID)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, DriverActionResponse{
		BookingID: bookingID.String(),
		Status:    models.BookingStatusAccepted.String(),
		Message:   "You have successfully accepted the ride",
	})
}

// CancelBooking - POST /v1/driver/bookings/{bookingId}/cancel
func (h *DriverHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Get Booking ID from URL params
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// 3. Call Service to cancel booking
	err = h.bookingService.CancelBooking(r.Context(), account.ID, bookingID)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, DriverActionResponse{
		BookingID: bookingID.String(),
		Status:    models.BookingStatusCancelled.String(),
		Message:   "Ride has been cancelled",
	})
}

// StartRide - POST /v1/driver/bookings/{bookingId}/start
func (h *DriverHandler) StartRide(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Get Booking ID from URL params
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// 3. Parse request body to get OTP code
	var req struct {
		OTP string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 4. Call Service to start ride
	err = h.bookingService.StartRide(r.Context(), account.ID, bookingID, req.OTP)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, DriverActionResponse{
		BookingID: bookingID.String(),
		Status:    models.BookingStatusStarted.String(),
		Message:   "Ride started successfully",
	})
}

// EndRide - POST /v1/driver/bookings/{bookingId}/end
func (h *DriverHandler) EndRide(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Get Booking ID from URL params
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// 3. Call Service to end ride
	err = h.bookingService.EndRide(r.Context(), account.ID, bookingID)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, DriverActionResponse{
		BookingID: bookingID.String(),
		Status:    models.BookingStatusCompleted.String(),
		Message:   "Ride completed successfully",
	})
}

// ListPendingRides - GET /v1/driver/bookings/pending
func (h *DriverHandler) ListPendingRides(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Parse Pagination (Page/Size -> Limit/Offset)
	// Example: ?page=2&page_size=10 -> limit=10, offset=10
	limit, offset := helper.GetPaginationParams(r)

	// 2. Call Service to get pending bookings
	bookings, err := h.bookingService.GetPendingRides(r.Context(), account.ID, limit, offset)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Respond with bookings
	var resp []CreateBookingResponse
	for _, booking := range bookings {
		resp = append(resp, CreateBookingResponse{
			ID:         booking.ID.String(),
			Status:     booking.Status,
			PickupLat:  booking.PickupLatitude,
			PickupLon:  booking.PickupLongitude,
			DropoffLat: booking.DropoffLatitude,
			DropoffLon: booking.DropoffLongitude,
			CreatedAt:  booking.CreatedAt,
			UpdatedAt:  booking.UpdatedAt,
		})
	}

	helper.RespondWithJSON(w, http.StatusOK, resp)
}

type ToggleAvailabilityRequest struct {
	Available bool `json:"available"`
}

// ToggleAvailability - PATCH /v1/driver/availability
func (h *DriverHandler) ToggleAvailability(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account
	account, err := GetAccountFromContext(r.Context())
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Parse Request
	var req ToggleAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 3. Call Service
	err = h.bookingService.ToggleDriverAvailability(r.Context(), account.ID, req.Available)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Respond
	statusMsg := "offline"
	if req.Available {
		statusMsg = "online"
	}
	helper.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Driver is now " + statusMsg,
	})
}
