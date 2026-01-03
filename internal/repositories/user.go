package repositories

import (
	"CabBookingService/internal/models"

	"github.com/google/uuid"
)

// UserRepository defines the methods for interacting with user data.
// By using an interface, our services can be tested with mocks
// and we can easily swap GORM for another DB if needed.
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
}
