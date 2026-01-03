package models

// Booking represents a booking
//type Booking struct {
//	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
//	CreatedAt time.Time
//	UpdatedAt time.Time
//	DeletedAt gorm.DeletedAt `gorm:"index"`
//
//	PassengerID uuid.UUID `gorm:"type:uuid;not null"`
//	Passenger   User      `gorm:"foreignKey:PassengerID"`
//
//	DriverID *uuid.UUID `gorm:"type:uuid"` // Pointer allows null (no driver yet)
//	Driver   *User      `gorm:"foreignKey:DriverID"`
//
//	Status BookingStatus `gorm:"size:50;not null"`
//
//	PickupLatitude   float64 `gorm:"not null"`
//	PickupLongitude  float64 `gorm:"not null"`
//	DropoffLatitude  float64 `gorm:"not null"`
//	DropoffLongitude float64 `gorm:"not null"`
//}
//
//func (*Booking) TableName() string {
//	return "booking"
//}
