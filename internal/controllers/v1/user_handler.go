package v1

import (
	"CabBookingService/internal/config"
	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserHandler holds the dependencies for the user controllers
type UserHandler struct {
	userRepo repositories.UserRepository
	cfg      *config.Config
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(cfg *config.Config, userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// RegisterRequest defines the expected JSON body for registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse defines the JSON response for a successful registration
type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// RegisterPassenger is the HTTP handler for POST /v1/register/passenger
func (h *UserHandler) RegisterPassenger(w http.ResponseWriter, r *http.Request) {
	// 1. Parse and validate the request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 2. Validate input
	if req.Username == "" || req.Password == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}
	// --- Business Logic is now in the Handler ---

	// 3. Check if username already exists
	_, err := h.userRepo.GetByUsername(req.Username)
	if err == nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Username already exists")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		helper.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// 4. Hash the password
	hashedPassword := req.Password // TODO: Hash the password!

	// 5. Create the user model
	user := &models.User{
		ID:          uuid.New(),
		Username:    req.Username,
		Password:    hashedPassword,
		IsPassenger: true,
		IsDriver:    false,
		IsAdmin:     false,
	}

	// 6. Save the user to the database
	if err := h.userRepo.Create(user); err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	// 7. Respond with the created user details
	resp := RegisterResponse{
		ID:       user.ID.String(),
		Username: user.Username,
	}
	helper.RespondWithJSON(w, http.StatusCreated, resp)
}

// --- Login Structs and Method ---

// LoginRequest defines the expected JSON body for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the JSON response for a successful login
type LoginResponse struct {
	Token string `json:"token"`
}

// Login is the HTTP handler for POST /v1/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Parse and validate the request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 2. Validate input
	if req.Username == "" || req.Password == "" {
		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// 3. Retrieve the user by username
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		helper.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// 4. Check password (TODO: Implement proper password hashing and comparison)
	hashedPassword := user.Password
	if req.Password != hashedPassword {
		helper.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// 5. Generate a token (TODO: Implement JWT or similar token generation)
	token := fmt.Sprintf("dummy-token-for-user-%s", user.Username)

	// 6. Respond with the token
	helper.RespondWithJSON(w, http.StatusOK, LoginResponse{Token: token})
}
