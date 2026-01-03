package services

import (
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/util"
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, booking *models.Booking) error
}

type paymentService struct {
	paymentRepo repositories.PaymentRepository
}

func NewPaymentService(paymentRepo repositories.PaymentRepository) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
	}
}

func (s *paymentService) ProcessPayment(ctx context.Context, booking *models.Booking) error {
	// 1. Calculate Amount (Mock Logic)
	// Real world: Calculate based on Distance + Time + Surge
	// Let's assume $5 Base + Distance * $2
	distance := util.DistanceKm(booking.PickupLatitude, booking.PickupLongitude, booking.DropoffLatitude, booking.DropoffLongitude)
	amount := 5.0 + (distance * 2.0)

	// 2. Get Payment Gateway (Default to "Cash" or "Stripe" seed data)
	gateway, err := s.paymentRepo.GetGatewayByName(ctx, "Stripe")
	if err != nil {
		// TODO: Fallback to Cash or other gateway
		return err
	}

	receipt := &models.PaymentReceipt{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		BookingId:        booking.ID,
		PaymentGatewayID: gateway.ID,
		Amount:           amount,
		Currency:         "USD",
		Details:          "Payment processed successfully",
	}

	return s.paymentRepo.CreateReceipt(ctx, receipt)
}
