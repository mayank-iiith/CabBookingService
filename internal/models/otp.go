package models

import (
	"time"
)

type OTP struct {
	BaseModel
	Code         string `gorm:"not null"`
	SentToNumber string `gorm:"not null"`
	ExpiresAt    time.Time
}

func (*OTP) TableName() string {
	return "otps"
}
