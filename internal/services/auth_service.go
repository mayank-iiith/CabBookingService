package services

import (
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

type AuthService interface {
	RegisterPassenger(username, password, name, phone string) (*models.Passenger, error)
	RegisterDriver(username, password, name, phone, plate, carModel string) (*models.Driver, error)
	Login(username, password string) (string, error)
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

func (a authService) RegisterPassenger(username, password, name, phoneNumber string) (*models.Passenger, error) {
	// 1. Check if username already exists
	if _, err := a.accountRepo.GetByUsername(username); err == nil {
		return nil, ErrUsernameAlreadyExists
	}

	// 2. Fetch Roles
	role, err := a.roleRepo.GetByName("ROLE_PASSENGER")
	if err != nil {
		return nil, errors.New("system error: passenger role not found")
	}

	// 3. Hash Password
	hashedPassword, err := GenerateHashFromPassword(password)
	if err != nil {
		return nil, err
	}

	// 4. Prepare Models
	now := time.Now().UTC()

	account := models.Account{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username: username,
		Password: hashedPassword,
		Roles:    []models.Role{*role},
	}

	passenger := models.Passenger{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		AccountId:   account.ID,
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	// 5. Transactionally save Account and Passenger
	err = a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		if err := tx.Create(&passenger).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	passenger.Account = account
	return &passenger, nil
}

func (a authService) RegisterDriver(username, password, name, phone, plate, carModel string) (*models.Driver, error) {
	// 1. Check if username already exists
	if _, err := a.accountRepo.GetByUsername(username); err == nil {
		return nil, ErrUsernameAlreadyExists
	}

	// 2. Fetch Roles
	role, err := a.roleRepo.GetByName("ROLE_DRIVER")
	if err != nil {
		return nil, errors.New("system error: driver role not found")
	}

	// 3. Hash Password
	hashedPassword, err := GenerateHashFromPassword(password)
	if err != nil {
		return nil, err
	}

	// 4. Prepare Models
	now := time.Now().UTC()

	account := models.Account{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username: username,
		Password: hashedPassword,
		Roles:    []models.Role{*role},
	}

	driver := models.Driver{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		AccountId:   account.ID,
		Name:        name,
		PhoneNumber: phone,
	}

	car := models.Car{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		DriverId:      driver.ID,
		PlateNumber:   plate,
		BrandAndModel: carModel,
	}

	// 5. Transactionally save Account, Driver, and Car
	err = a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		if err := tx.Create(&driver).Error; err != nil {
			return err
		}
		if err := tx.Create(&car).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	driver.Account = account
	driver.Car = car
	return &driver, nil
}

func (a authService) Login(username, password string) (string, error) {
	account, err := a.accountRepo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	if err := CompareHashAndPassword(account.Password, password); err != nil {
		return "", errors.New("invalid username or password")
	}

	// Build Roles Slice
	roles := make([]string, len(account.Roles))
	for _, role := range account.Roles {
		roles = append(roles, role.Name)
	}

	// Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   account.ID.String(),
		"roles": roles,
		"exp":   time.Now().Add(time.Duration(a.jwtExpiry) * time.Second).Unix(),
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateHashFromPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CompareHashAndPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
