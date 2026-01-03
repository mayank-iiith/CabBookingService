package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverRepository interface {
	Create(driver *models.Driver) error
	GetByID(id uuid.UUID) (*models.Driver, error)
	GetByAccountID(accountID uuid.UUID) (*models.Driver, error)
	UpdateAvailability(driverID uuid.UUID, isAvailable bool) error
}

type gormDriverRepository struct {
	db *gorm.DB
}

func NewGormDriverRepository(db *gorm.DB) DriverRepository {
	return &gormDriverRepository{db: db}
}

func (g *gormDriverRepository) Create(driver *models.Driver) error {
	return g.db.Create(driver).Error
}

func (g *gormDriverRepository) GetByID(id uuid.UUID) (*models.Driver, error) {
	var driver models.Driver
	err := g.db.
		Where("id = ?", id).
		Preload("Account").
		Preload("Car").
		First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (g *gormDriverRepository) GetByAccountID(accountID uuid.UUID) (*models.Driver, error) {
	var driver models.Driver
	err := g.db.
		Where("account_id = ?", accountID).
		Preload("Account").
		Preload("Car").
		First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (g *gormDriverRepository) UpdateAvailability(id uuid.UUID, isAvailable bool) error {
	return g.db.Model(&models.Driver{}).
		Where("id = ?", id).
		Update("is_available", isAvailable).Error
}
