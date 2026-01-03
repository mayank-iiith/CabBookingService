package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(ctx context.Context, otp *models.OTP) error
	GetById(ctx context.Context, id uuid.UUID) (*models.OTP, error)
}

type gormOTPRepository struct {
	db *gorm.DB
}

func NewGormOTPRepository(db *gorm.DB) OTPRepository {
	return &gormOTPRepository{db: db}
}

func (r *gormOTPRepository) Create(ctx context.Context, otp *models.OTP) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(otp).Error
}

func (r *gormOTPRepository) GetById(ctx context.Context, id uuid.UUID) (*models.OTP, error) {
	tx := db.NewGormTx(ctx, r.db)

	var otp models.OTP
	if err := tx.First(&otp, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &otp, nil
}
