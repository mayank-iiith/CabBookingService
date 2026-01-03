package models

import (
	"github.com/google/uuid"
)

type PaymentGateway struct {
	BaseModel
	Name string `gorm:"not null;unique"` // e.g., "Stripe", "PayPal"
}

func (*PaymentGateway) TableName() string {
	return "payment_gateways"
}

type PaymentReceipt struct {
	BaseModel

	BookingId uuid.UUID `gorm:"type:uuid;not null;unique"` // One receipt per booking
	Booking   Booking   `gorm:"foreignKey:BookingId"`

	PaymentGatewayID uuid.UUID      `gorm:"type:uuid;not null"`
	PaymentGateway   PaymentGateway `gorm:"foreignKey:PaymentGatewayID"`

	Amount   float64 `gorm:"not null"`
	Currency string  `gorm:"default:'USD'"`
	Details  string  `gorm:"type:text"` // JSON dump from gateway
}

func (*PaymentReceipt) TableName() string {
	return "payment_receipts"
}
