package models

import (
	"time"

	"github.com/google/uuid"
)

type Driver struct {
	BaseModel

	AccountId uuid.UUID `gorm:"type:uuid;not null"`
	Account   Account   `gorm:"foreignKey:AccountId"`

	Name           string
	Gender         *Gender
	DateOfBirth    *time.Time `gorm:"type:date"` // Maps to SQL DATE
	PhoneNumber    string
	LicenseDetails string
	IsAvailable    bool
	ActiveCity     string

	// Has-One relationship with Car
	Car Car `gorm:"foreignKey:DriverId"`

	// Rating
	AverageRating float64 `gorm:"default:0.0"`
	RatingCount   int     `gorm:"default:0"`

	LastKnownLatitude  *float64
	LastKnownLongitude *float64
	// Helper struct for Go logic, not GORM
	LastKnownLocation *ExactLocation `gorm:"-"`
}

type ExactLocation struct {
	Latitude  float64
	Longitude float64
}

func (*Driver) TableName() string {
	return "drivers"
}
