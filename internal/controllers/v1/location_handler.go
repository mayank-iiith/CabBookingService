package v1

import (
	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/models"
	"CabBookingService/internal/services"
	"encoding/json"
	"net/http"
)

type LocationHandler struct {
	locationService services.LocationService
}

func NewLocationHandler(locationService services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *LocationHandler) UpdateDriverLocation(w http.ResponseWriter, r *http.Request) {
	// 1. Get Account from context (set by AuthMiddleware)
	account, ok := r.Context().Value(AccountKey).(*models.Account)
	if !ok {
		helper.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Parse request body
	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 3. Update location via service
	err := h.locationService.UpdateDriverLocation(r.Context(), account.ID, req.Latitude, req.Longitude)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Respond with success
	helper.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Location updated successfully"})
}
