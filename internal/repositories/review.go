package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
}

type gormReviewRepository struct {
	db *gorm.DB
}

func NewGormReviewRepository(db *gorm.DB) ReviewRepository {
	return &gormReviewRepository{db: db}
}

func (r *gormReviewRepository) Create(ctx context.Context, review *models.Review) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(review).Error
}
