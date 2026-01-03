package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user (can be a passenger or a driver)
type User struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Username string
	Password string

	// Role management
	IsPassenger bool
	IsDriver    bool
	IsAdmin     bool
}

func (user *User) TableName() string {
	return "users"
}
