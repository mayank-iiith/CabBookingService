package v1

import (
	"CabBookingService/internal/config"
	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/services"
	"encoding/json"
	"net/http"
)

// UserHandler handles authentication requests
type UserHandler struct {
	authService services.AuthService
	cfg         *config.Config
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(cfg *config.Config, authService services.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// --- Requests / Responses ---

type RegisterPassengerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type RegisterDriverRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	PlateNumber string `json:"plate_number"`
	CarModel    string `json:"car_model"`
}

// RegisterResponse defines the JSON response for a successful registration
type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// LoginRequest defines the expected JSON body for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the JSON response for a successful login
type LoginResponse struct {
	Token string `json:"token"`
}

// --- Handlers ---

// RegisterPassenger : POST /v1/register/passenger
func (h *UserHandler) RegisterPassenger(w http.ResponseWriter, r *http.Request) {
	var req RegisterPassengerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Call Service
	passenger, err := h.authService.RegisterPassenger(r.Context(), req.Username, req.Password, req.Name, req.PhoneNumber)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusCreated, RegisterResponse{
		ID:       passenger.ID.String(),
		Username: passenger.Account.Username,
	})
}

// RegisterDriver : POST /v1/register/driver
func (h *UserHandler) RegisterDriver(w http.ResponseWriter, r *http.Request) {
	var req RegisterDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Call Service
	driver, err := h.authService.RegisterDriver(r.Context(), req.Username, req.Password, req.Name, req.PhoneNumber, req.PlateNumber, req.CarModel)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusCreated, RegisterResponse{
		ID:       driver.ID.String(),
		Username: driver.Account.Username,
	})
}

// Login : POST /v1/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Call Service
	token, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, LoginResponse{Token: token})
}
