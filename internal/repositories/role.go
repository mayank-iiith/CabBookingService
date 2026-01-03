package repositories

import (
	"CabBookingService/internal/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetByName(name string) (*models.Role, error)
}

type gormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepository{db: db}
}

func (g *gormRoleRepository) GetByName(name string) (*models.Role, error) {
	var role models.Role
	// We use "name = ?" to find the role (e.g., "ROLE_PASSENGER")
	err := g.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
