package services

import (
	"context"
	"errors"
	"time"

	"CabBookingService/internal/db"
	"CabBookingService/internal/domain"
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameAlreadyExists     = errors.New("username already exists")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)

// UserClaims extends standard JWT claims with application-specific data.
// By embedding jwt.RegisteredClaims, we gain automatic JSON mapping for standard fields (exp, iss, sub, etc.)
// and helper methods like GetExpirationTime().
type UserClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type AuthService interface {
	RegisterPassenger(ctx context.Context, username, password, name, phone string) (*models.Passenger, error)
	RegisterDriver(ctx context.Context, username, password, name, phone, plate, carModel string) (*models.Driver, error)
	Login(ctx context.Context, username, password string) (string, error)
}

type authService struct {
	accountRepo   repositories.AccountRepository
	passengerRepo repositories.PassengerRepository
	driverRepo    repositories.DriverRepository
	roleRepo      repositories.RoleRepository

	db        *gorm.DB
	jwtSecret string
	jwtExpiry int64
}

func NewAuthService(
	accountRepo repositories.AccountRepository,
	passengerRepo repositories.PassengerRepository,
	driverRepo repositories.DriverRepository,
	roleRepo repositories.RoleRepository,
	db *gorm.DB,
	jwtSecret string,
	jwtExpiry int64,
) AuthService {
	return &authService{
		accountRepo:   accountRepo,
		passengerRepo: passengerRepo,
		driverRepo:    driverRepo,
		roleRepo:      roleRepo,
		db:            db,
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
	}
}

func (a *authService) RegisterPassenger(ctx context.Context, username, password, name, phoneNumber string) (*models.Passenger, error) {
	// 1. Fetch Role
	role, err := a.roleRepo.GetByName(ctx, domain.RolePassenger)
	if err != nil {
		return nil, errors.New("system error: passenger role not found")
	}

	// 2. Check for Existing Account, We preload roles here to ensure the duplicate check works
	account, err := a.accountRepo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Flag to track logic
	isExistingAccount := err == nil

	now := time.Now().UTC()

	// Validations before starting transaction
	if isExistingAccount {
		// A.1 Verify Password
		if err := CompareHashAndPassword(account.Password, password); err != nil {
			return nil, errors.New("username exists: incorrect password provided to link account")
		}
		// A.2 Check Duplicate Role
		for _, r := range account.Roles {
			if r.Name == domain.RolePassenger {
				// They are already a passenger, we can't register them again
				return nil, errors.New("passenger is already registered as a passenger")
			}
		}
	} else {
		// B.1 Prepare New Account
		hashedPassword, err := GenerateHashFromPassword(password)
		if err != nil {
			return nil, err
		}
		account = &models.Account{
			BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
			Username:  username,
			Password:  hashedPassword,
			Roles:     []models.Role{*role},
		}
	}

	// 3. Prepare Passenger Profile
	passenger := &models.Passenger{
		BaseModel:   models.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
		AccountId:   account.ID,
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	// 4. Execute Atomic Transaction
	// We use the transaction 'tx' for ALL writes to ensure either everything saves or nothing saves.
	err = db.NewGormTx(ctx, a.db).Transaction(func(tx *gorm.DB) error {
		if isExistingAccount {
			// Link the Role to the existing Account
			if err := tx.Model(account).Association("Roles").Append(role); err != nil {
				return err
			}
		} else {
			// Create the new Account
			if err := tx.Create(account).Error; err != nil {
				return err
			}
		}

		// Create the Passenger Profile
		if err := tx.Create(passenger).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to register passenger")
		return nil, err
	}

	// Populate account for return
	passenger.Account = *account
	log.Info().Str("username", username).Msg("New passenger registered")
	return passenger, nil
}

func (a *authService) RegisterDriver(ctx context.Context, username, password, name, phone, plate, carModel string) (*models.Driver, error) {
	// 1. Fetch Role
	role, err := a.roleRepo.GetByName(ctx, domain.RoleDriver)
	if err != nil {
		return nil, errors.New("system error: driver role not found")
	}

	// 2. Check for Existing Account, We preload roles here to ensure the duplicate check works
	account, err := a.accountRepo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Flag to track logic
	isExistingAccount := err == nil

	now := time.Now().UTC()

	// Validations before starting transaction
	if isExistingAccount {
		// A.1 Verify Password
		if err := CompareHashAndPassword(account.Password, password); err != nil {
			return nil, errors.New("username exists: incorrect password provided to link account")
		}
		// A.2 Check Duplicate Role
		for _, r := range account.Roles {
			if r.Name == domain.RoleDriver {
				// They are already a driver, we can't register them again
				return nil, errors.New("user is already registered as a driver")
			}
		}
	} else {
		// B.1 Prepare New Account
		hashedPassword, err := GenerateHashFromPassword(password)
		if err != nil {
			return nil, err
		}
		account = &models.Account{
			BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
			Username:  username,
			Password:  hashedPassword,
			Roles:     []models.Role{*role},
		}
	}

	// 3. Prepare Passenger Profile
	driver := &models.Driver{
		BaseModel:   models.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
		AccountId:   account.ID,
		Name:        name,
		PhoneNumber: phone,
	}

	// 4. Prepare Car Profile
	car := models.Car{
		BaseModel:     models.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
		DriverId:      driver.ID,
		PlateNumber:   plate,
		BrandAndModel: carModel,
	}

	// 5. Execute Atomic Transaction
	// We use the transaction 'tx' for ALL writes to ensure either everything saves or nothing saves.
	err = db.NewGormTx(ctx, a.db).Transaction(func(tx *gorm.DB) error {
		// Step 5a: Handle Account
		if isExistingAccount {
			// Link the Role to the existing Account
			if err := tx.Model(account).Association("Roles").Append(role); err != nil {
				return err
			}
		} else {
			// Create the new Account
			if err := tx.Create(account).Error; err != nil {
				return err
			}
		}

		// Step 5b: Create Driver
		if err := tx.Create(driver).Error; err != nil {
			return err
		}

		// Step 5c: Create Car
		if err := tx.Create(car).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to register driver")
		return nil, err
	}

	// Populate fields for return
	driver.Account = *account
	driver.Car = car

	log.Info().Str("username", username).Msg("New driver registered")
	return driver, nil
}

func (a *authService) Login(ctx context.Context, username, password string) (string, error) {
	account, err := a.accountRepo.GetByUsername(ctx, username)
	if err != nil {
		// Log failed login attempt (Security Audit)
		log.Error().Err(err).Str("username", username).Msg("Login failed: user not found")
		return "", ErrInvalidUsernameOrPassword
	}

	if err := CompareHashAndPassword(account.Password, password); err != nil {
		// Log failed login attempt
		log.Error().Err(err).Str("username", username).Msg("Login failed: invalid password")
		return "", ErrInvalidUsernameOrPassword
	}

	// 3. Build Roles Slice
	roles := make([]string, 0, len(account.Roles))
	for _, role := range account.Roles {
		roles = append(roles, role.Name)
	}

	// 4. Create Claims using jwt.RegisteredClaims
	now := time.Now().UTC()
	expirationTime := now.Add(time.Duration(a.jwtExpiry) * time.Second)

	claims := UserClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			// Standard fields as per RFC 7519
			Issuer:    "cab-booking-service",
			Subject:   account.ID.String(),
			Audience:  jwt.ClaimStrings{"cab-booking-api"},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(), // JTI: Unique identifier for this token
		},
	}

	// 5. Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to sign JWT token")
		return "", err
	}

	log.Info().
		Str("user_id", account.ID.String()).
		Str("username", username).
		Msg("User logged in successfully")
	return tokenString, nil
}

func GenerateHashFromPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CompareHashAndPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
