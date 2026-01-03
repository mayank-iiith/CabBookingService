package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PassengerRepository interface {
	Create(passenger *models.Passenger) error
	GetByID(id uuid.UUID) (*models.Passenger, error)
	GetByAccountID(accountID uuid.UUID) (*models.Passenger, error)
}

type gormPassengerRepository struct {
	db *gorm.DB
}

func NewGormPassengerRepository(db *gorm.DB) PassengerRepository {
	return &gormPassengerRepository{db: db}
}

func (g *gormPassengerRepository) Create(passenger *models.Passenger) error {
	return g.db.Create(passenger).Error
}

func (g *gormPassengerRepository) GetByID(id uuid.UUID) (*models.Passenger, error) {
	var passenger models.Passenger
	// Preload Account to get username/email
	err := g.db.
		Where("id = ?", id).
		Preload("Account").
		First(&passenger).Error
	if err != nil {
		return nil, err
	}
	return &passenger, nil
}

func (g *gormPassengerRepository) GetByAccountID(accountID uuid.UUID) (*models.Passenger, error) {
	var passenger models.Passenger
	err := g.db.Where("account_id = ?", accountID).Preload("Account").First(&passenger).Error
	if err != nil {
		return nil, err
	}
	return &passenger, nil
}
