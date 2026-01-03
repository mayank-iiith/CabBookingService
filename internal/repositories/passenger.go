package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PassengerRepository interface {
	Create(ctx context.Context, passenger *models.Passenger) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Passenger, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*models.Passenger, error)
}

type gormPassengerRepository struct {
	db *gorm.DB
}

func NewGormPassengerRepository(db *gorm.DB) PassengerRepository {
	return &gormPassengerRepository{db: db}
}

func (r *gormPassengerRepository) Create(ctx context.Context, passenger *models.Passenger) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(passenger).Error
}

func (r *gormPassengerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Passenger, error) {
	tx := db.NewGormTx(ctx, r.db)

	var passenger models.Passenger
	// Preload Account to get username/email
	err := tx.Where("id = ?", id).
		Preload("Account").
		First(&passenger).Error
	if err != nil {
		return nil, err
	}
	return &passenger, nil
}

func (r *gormPassengerRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*models.Passenger, error) {
	tx := db.NewGormTx(ctx, r.db)

	var passenger models.Passenger
	err := tx.Where("account_id = ?", accountID).
		Preload("Account").
		First(&passenger).Error
	if err != nil {
		return nil, err
	}
	return &passenger, nil
}
