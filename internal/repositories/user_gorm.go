package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormUserRepository is the GORM implementation of the UserRepository
type gormUserRepository struct {
	db *gorm.DB // The GORM database connection
}

// Ensure gormUserRepository implements UserRepository
var _ UserRepository = &gormUserRepository{nil}

// NewGormUserRepository creates a new GORM user repository
func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

// Create saves a new user to the database
func (r *gormUserRepository) Create(user *models.User) error {
	// GORM's Create method will run the BeforeCreate hook
	// in your model, generate the UUID, and save the user.
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a user by their UUID
func (r *gormUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by their username
func (r *gormUserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
