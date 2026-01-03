package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(otp *models.OTP) error
	GetById(id uuid.UUID) (*models.OTP, error)
}

type gormOTPRepository struct {
	db *gorm.DB
}

func NewGormOTPRepository(db *gorm.DB) OTPRepository {
	return &gormOTPRepository{db: db}
}

func (r *gormOTPRepository) Create(otp *models.OTP) error {
	return r.db.Create(otp).Error
}

func (r *gormOTPRepository) GetById(id uuid.UUID) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.First(&otp, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &otp, nil
}
