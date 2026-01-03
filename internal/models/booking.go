package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	BaseModel

	PassengerId uuid.UUID `gorm:"type:uuid;not null"`
	Passenger   Passenger `gorm:"foreignKey:PassengerId"` // Links to Passenger Profile

	DriverId *uuid.UUID `gorm:"type:uuid;not null"` // Pointer allows null (no driver yet)
	Driver   *Driver    `gorm:"foreignKey:DriverId"`

	Status BookingStatus

	// --- NEW: Security Feature ---
	// This stores which drivers were offered this ride. Only drivers in this list can accept it.
	// gorm: "many2many" tells GORM to use the join table we just created.
	NotifiedDrivers []Driver `gorm:"many2many:booking_notified_drivers;"`

	// --- OTP Relation ---
	RideStartOTPId *uuid.UUID `gorm:"type:uuid"`
	RideStartOTP   *OTP       `gorm:"foreignKey:RideStartOTPId"`

	PickupLatitude   float64 `gorm:"not null"`
	PickupLongitude  float64 `gorm:"not null"`
	DropoffLatitude  float64 `gorm:"not null"`
	DropoffLongitude float64 `gorm:"not null"`

	// Reviews
	ReviewByPassengerId *uuid.UUID `gorm:"type:uuid"`
	ReviewByPassenger   *Review    `gorm:"foreignKey:ReviewByPassengerId"`

	ReviewByDriverId *uuid.UUID `gorm:"type:uuid"`
	ReviewByDriver   *Review    `gorm:"foreignKey:ReviewByDriverId"`

	ScheduledTime *time.Time // Nullable for immediate rides
}

func (*Booking) TableName() string {
	return "bookings"
}
