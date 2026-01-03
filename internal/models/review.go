package models

import (
	"github.com/google/uuid"
)

type Review struct {
	BaseModel

	Rating int    `gorm:"not null;check:rating >= 1 AND rating <= 5"` // 1-5 Stars
	Note   string `gorm:"type:text"`

	// Foreign Key to Booking Id
	DriverID    *uuid.UUID `gorm:"type:uuid;not null"`
	PassengerId *uuid.UUID `gorm:"type:uuid;not null"`
	BookingId   uuid.UUID  `gorm:"type:uuid;not null"`
}

func (*Review) TableName() string {
	return "reviews"
}

//func (r *Review) BeforeSave() error {
//	if r.Rating < 1 || r.Rating > 5 {
//		return errors.New("invalid rating")
//	}
//
//	if r.DriverID == nil && r.PassengerId == nil {
//		return errors.New("either DriverID or PassengerId must be set")
//	}
//
//	return nil
//}
