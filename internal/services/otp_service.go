package services

import (
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type OTPService interface {
	GenerateOTP(ctx context.Context, phoneNumber string) (*models.OTP, error)
	ValidateOTP(ctx context.Context, otpID uuid.UUID, code string) bool
}

type otpService struct {
	otpRepo repositories.OTPRepository
}

func NewOTPService(otpRepo repositories.OTPRepository) OTPService {
	return &otpService{otpRepo: otpRepo}
}

func (s *otpService) GenerateOTP(ctx context.Context, phoneNumber string) (*models.OTP, error) {
	// Simple 4-digit random code
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := strconv.Itoa(1000 + r.Intn(9000))

	now := time.Now()

	otp := &models.OTP{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Code:         code,
		SentToNumber: phoneNumber,
		ExpiresAt:    now.Add(60 * time.Minute),
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return nil, err
	}

	return otp, nil
}

func (s *otpService) ValidateOTP(ctx context.Context, otpID uuid.UUID, code string) bool {
	otp, err := s.otpRepo.GetById(ctx, otpID)
	if err != nil {
		return false
	}

	if otp.Code != code {
		return false
	}

	if time.Now().After(otp.ExpiresAt) {
		return false
	}

	return true
}
