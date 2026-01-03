package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*models.Role, error)
}

type gormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepository{db: db}
}

func (r *gormRoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	tx := db.NewGormTx(ctx, r.db)

	var role models.Role
	// We use "name = ?" to find the role (e.g., "ROLE_PASSENGER")
	err := tx.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
