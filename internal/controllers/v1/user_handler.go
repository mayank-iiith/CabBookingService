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

// RegisterPassenger godoc
// @Summary      Register a new passenger
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterPassengerRequest true "Passenger Data"
// @Success      201  {object}  RegisterResponse
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
	passenger, err := h.authService.RegisterPassenger(req.Username, req.Password, req.Name, req.PhoneNumber)
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
	driver, err := h.authService.RegisterDriver(req.Username, req.Password, req.Name, req.PhoneNumber, req.PlateNumber, req.CarModel)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusCreated, RegisterResponse{
		ID:       driver.ID.String(),
		Username: driver.Account.Username,
	})
}

// Login godoc
// @Summary      Login and get JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Credentials"
// @Success      200  {object}  LoginResponse
// @Router       /login [post]
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
	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		helper.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// RegisterPassengerOld is the HTTP handler for POST /v1/register/passenger
//func (h *UserHandler) RegisterPassengerOld(w http.ResponseWriter, r *http.Request) {
//	// 1. Parse and validate the request
//	var req RegisterRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
//		return
//	}
//
//	// 2. Validate input
//	if req.Username == "" || req.Password == "" {
//		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
//		return
//	}
//	// --- Business Logic is now in the Handler ---
//
//	// 3. Check if username already exists
//	_, err := h.userRepo.GetByUsername(req.Username)
//	if err == nil {
//		helper.RespondWithError(w, http.StatusBadRequest, "Username already exists")
//		return
//	}
//	if !errors.Is(err, gorm.ErrRecordNotFound) {
//		helper.RespondWithError(w, http.StatusInternalServerError, "Database error")
//		return
//	}
//
//	// 4. Hash the password
//	hashedPassword, err := GenerateHashFromPassword(req.Password)
//	if err != nil {
//		helper.RespondWithError(w, http.StatusInternalServerError, "Could not hash password")
//		return
//	}
//
//	// 5. Create the user model
//	user := &models.User{
//		ID:          uuid.New(),
//		Username:    req.Username,
//		Password:    hashedPassword,
//		IsPassenger: true,
//		IsDriver:    false,
//		IsAdmin:     false,
//	}
//
//	// 6. Save the user to the database
//	if err := h.userRepo.Create(user); err != nil {
//		helper.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
//		return
//	}
//
//	// 7. Respond with the created user details
//	resp := RegisterResponse{
//		ID:       user.ID.String(),
//		Username: user.Username,
//	}
//	helper.RespondWithJSON(w, http.StatusCreated, resp)
//}

// --- LoginOld Structs and Method ---

// LoginOld is the HTTP handler for POST /v1/login
//func (h *UserHandler) LoginOld(w http.ResponseWriter, r *http.Request) {
//	// 1. Parse and validate the request
//	var req LoginRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		helper.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
//		return
//	}
//
//	// 2. Validate input
//	if req.Username == "" || req.Password == "" {
//		helper.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
//		return
//	}
//
//	// 3. Retrieve the user by username
//	user, err := h.userRepo.GetByUsername(req.Username)
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			helper.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
//			return
//		}
//		helper.RespondWithError(w, http.StatusInternalServerError, "Database error")
//		return
//	}
//
//	// 4. Check password
//	err = CompareHashAndPassword(user.Password, req.Password)
//	if err != nil {
//		helper.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
//		return
//	}
//
//	// 5. Create JWT token
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//		"sub": user.ID.String(),
//		"rol": map[string]bool{
//			"passenger": user.IsPassenger,
//			"driver":    user.IsDriver,
//			"admin":     user.IsAdmin,
//		},
//		"exp": time.Now().Add(time.Duration(h.cfg.JWTExpiresIn) * time.Second).Unix(),
//		"iat": time.Now().Unix(),
//	})
//
//	// 5. Sign the token
//	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
//	if err != nil {
//		helper.RespondWithError(w, http.StatusInternalServerError, "Could not generate token")
//		return
//	}
//
//	// 6. Respond with the token
//	helper.RespondWithJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
//}
//
//func GenerateHashFromPassword(password string) (string, error) {
//	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
//	return string(bytes), err
//}
//
//func CompareHashAndPassword(hash, password string) error {
//	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
//}
