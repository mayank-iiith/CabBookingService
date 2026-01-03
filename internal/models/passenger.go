package models

import (
	"time"

	"github.com/google/uuid"
)

type Passenger struct {
	BaseModel

	AccountId uuid.UUID `gorm:"type:uuid;not null"`
	Account   Account   `gorm:"foreignKey:AccountId"`

	Name        string
	PhoneNumber string
	Gender      *Gender
	DateOfBirth *time.Time `gorm:"type:date"` // 'type:date' forces Postgres to use the DATE column, ignoring the time part of time.Time

	// Rating
	AverageRating float64 `gorm:"default:0.0"`
	RatingCount   int     `gorm:"default:0"`
}

func (*Passenger) TableName() string {
	return "passengers"
}
