package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepository interface {
	GetByUsername(username string) (*models.Account, error)
	GetByID(id uuid.UUID) (*models.Account, error)

	// Create We don't necessarily need a separate Create here as AuthService handles it transactionally,
	// but it's good practice to have.
	Create(account *models.Account) error
}

type gormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) AccountRepository {
	return &gormAccountRepository{db: db}
}

func (g *gormAccountRepository) Create(account *models.Account) error {
	return g.db.Create(account).Error
}

func (g *gormAccountRepository) GetByUsername(username string) (*models.Account, error) {
	var account models.Account
	// Critical: Preload "Roles" so we know if they are a Passenger/Driver/Admin
	err := g.db.
		Where("username = ?", username).
		Preload("Roles").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (g *gormAccountRepository) GetByID(id uuid.UUID) (*models.Account, error) {
	var account models.Account
	// Critical: Preload "Roles" so we know if they are a Passenger/Driver/Admin
	err := g.db.
		Where("id = ?", id).
		Preload("Roles").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
