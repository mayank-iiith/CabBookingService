package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverRepository interface {
	Create(ctx context.Context, driver *models.Driver) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Driver, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*models.Driver, error)
	GetByAccountIDs(ctx context.Context, accountIDs []uuid.UUID) ([]models.Driver, error)
	UpdateAvailability(ctx context.Context, driverID uuid.UUID, isAvailable bool) error
	UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lon float64) error
}

type gormDriverRepository struct {
	db *gorm.DB
}

func NewGormDriverRepository(db *gorm.DB) DriverRepository {
	return &gormDriverRepository{db: db}
}

func (r *gormDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(driver).Error
}

func (r *gormDriverRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Driver, error) {
	tx := db.NewGormTx(ctx, r.db)

	var driver models.Driver
	err := tx.Where("id = ?", id).
		Preload("Account").
		Preload("Car").
		First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, err
}

func (r *gormDriverRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*models.Driver, error) {
	tx := db.NewGormTx(ctx, r.db)

	var driver models.Driver
	err := tx.Where("account_id = ?", accountID).
		Preload("Account").
		Preload("Car").
		First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *gormDriverRepository) GetByAccountIDs(ctx context.Context, accountIDs []uuid.UUID) ([]models.Driver, error) {
	tx := db.NewGormTx(ctx, r.db)

	var drivers []models.Driver
	err := tx.Where("account_id IN ?", accountIDs).
		Preload("Account").
		Preload("Car").
		Find(&drivers).Error
	if err != nil {
		return nil, err
	}
	return drivers, nil
}

func (r *gormDriverRepository) UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Model(&models.Driver{}).
		Where("id = ?", id).
		Update("is_available", isAvailable).Error
}

func (r *gormDriverRepository) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lon float64) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Model(&models.Driver{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"last_known_latitude":  lat,
			"last_known_longitude": lon,
		}).Error
}
