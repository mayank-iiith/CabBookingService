package repositories

import (
	"CabBookingService/internal/models"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
}

type gormReviewRepository struct {
	db *gorm.DB
}

func NewGormReviewRepository(db *gorm.DB) ReviewRepository {
	return &gormReviewRepository{db: db}
}

func (r *gormReviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}
