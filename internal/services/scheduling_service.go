package services

import (
	"context"
	"time"

	"CabBookingService/internal/domain"
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services/queue"

	"github.com/rs/zerolog/log"
)

type SchedulingService interface {
	Start(ctx context.Context)
}

type schedulingService struct {
	bookingRepo   repositories.BookingRepository
	messageQueue  queue.MessageQueue
	checkInterval time.Duration
	window        time.Duration
}

func NewSchedulingService(
	bookingRepo repositories.BookingRepository,
	messageQueue queue.MessageQueue,
) SchedulingService {
	return &schedulingService{
		bookingRepo:   bookingRepo,
		messageQueue:  messageQueue,
		checkInterval: 1 * time.Minute,  // Run every minute
		window:        15 * time.Minute, // Process rides 15 mins before time
	}
}

func (s schedulingService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	log.Info().Msg("Scheduling Service started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Info().Msg("Scheduling Service stopped")
				return
			case <-ticker.C:
				s.processScheduledBookings(ctx)
			}
		}
	}()
}

func (s schedulingService) processScheduledBookings(ctx context.Context) {
	// 1. Find bookings that are SCHEDULED and due within the window
	cutoff := time.Now().Add(s.window)

	bookings, err := s.bookingRepo.GetDueScheduledBookings(ctx, cutoff)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch scheduled bookings")
		return
	}

	for _, booking := range bookings {
		// 2. Update Status to REQUESTED
		if err := s.bookingRepo.UpdateStatus(ctx, booking.ID, models.BookingStatusRequested); err != nil {
			log.Error().Err(err).Str("booking_id", booking.ID.String()).Msg("Failed to activate scheduled booking")
			continue
		}

		// 3. Push to Matching Queue
		log.Info().Str("booking_id", booking.ID.String()).Msg("Activating scheduled booking")
		if err := s.messageQueue.Publish(ctx, domain.TopicDriverMatching, booking.ID); err != nil {
			log.Error().Err(err).Str("booking_id", booking.ID.String()).Msg("Failed to publish to matching queue")
		}

		log.Info().Str("booking_id", booking.ID.String()).Msg("Published scheduled booking to driver matching queue")
	}
}
