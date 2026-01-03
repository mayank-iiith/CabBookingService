package repositories

import (
	"CabBookingService/internal/models"
	"context"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	GetGatewayByName(ctx context.Context, name string) (*models.PaymentGateway, error)
	CreateReceipt(ctx context.Context, receipt *models.PaymentReceipt) error
}

type gormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) PaymentRepository {
	return &gormPaymentRepository{db: db}
}

func (r *gormPaymentRepository) GetGatewayByName(ctx context.Context, name string) (*models.PaymentGateway, error) {
	var gateway models.PaymentGateway
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&gateway).Error
	if err != nil {
		return nil, err
	}
	return &gateway, nil
}

func (r *gormPaymentRepository) CreateReceipt(ctx context.Context, receipt *models.PaymentReceipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}
