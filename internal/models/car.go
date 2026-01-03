package models

import (
	"github.com/google/uuid"
)

type Car struct {
	BaseModel

	DriverId      uuid.UUID `gorm:"type:uuid;not null;unique"`
	PlateNumber   string    `gorm:"uniqueIndex"`
	BrandAndModel string
	Color         string
	CarType       string
}

func (*Car) TableName() string {
	return "cars"
}
