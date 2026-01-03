package services

import (
	"CabBookingService/internal/domain"
	"CabBookingService/internal/services/filters"
	"context"
	"fmt"

	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services/queue"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DriverMatchingService defines the contract for driver matching services
type DriverMatchingService interface {
	StartConsuming() error
}

type driverMatchingService struct {
	queue           queue.MessageQueue
	locationService LocationService
	bookingRepo     repositories.BookingRepository
	driverRepo      repositories.DriverRepository
	filters         []filters.DriverFilter
}

func NewDriverMatchingService(
	queue queue.MessageQueue,
	locationService LocationService,
	bookingRepo repositories.BookingRepository,
	driverRepo repositories.DriverRepository,
) DriverMatchingService {
	return &driverMatchingService{
		queue:           queue,
		locationService: locationService,
		bookingRepo:     bookingRepo,
		driverRepo:      driverRepo,
		filters: []filters.DriverFilter{
			// Add filters here
			filters.NewETABasedFilter(5.0), // Max 5km away
			filters.NewGenderFilter(),
		},
	}
}

func (s *driverMatchingService) StartConsuming() error {
	ch, err := s.queue.Subscribe(domain.TopicDriverMatching)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", domain.TopicDriverMatching, err)
	}
	log.Info().Str("topic", domain.TopicDriverMatching).Msg("Subscribed to topic")

	// Start a background goroutine to consume messages
	// TODO: When service is closed we should also stop this goroutine gracefully
	go func() {
		log.Info().Msg("[DriverMatching] Started consuming messages...")
		for msg := range ch {
			// Type Assertion
			bookingID, ok := msg.(uuid.UUID)
			if !ok {
				log.Warn().Interface("msg", msg).Msg("[DriverMatching] Invalid message format received")
				continue
				// TODO: In production, consider dead-letter queue or alerting
			}
			s.handleDriverMatching(bookingID)
		}
	}()
	return nil
}

func (s *driverMatchingService) handleDriverMatching(bookingID uuid.UUID) {
	log.Info().Str("booking_id", bookingID.String()).Msg("Handling driver matching")

	ctx := context.Background() // TODO: Either pass context from message or create with timeout

	// 1. Fetch Booking
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Failed to fetch booking")
		return
	}

	// 2. Find nearby drivers using pickup location from booking
	radiusToSearch := 2.0 // in km // TODO: Make configurable
	nearbyDriverIDs := s.locationService.GetNearbyDrivers(booking.PickupLatitude, booking.PickupLongitude, radiusToSearch)
	if len(nearbyDriverIDs) == 0 {
		log.Info().Str("booking_id", bookingID.String()).Msg("No drivers nearby")
		return
	}

	// 3. Fetch Full Driver Profiles
	candidateDrivers, err := s.driverRepo.GetByAccountIDs(ctx, nearbyDriverIDs)
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Error fetching driver profiles")
		return
	}

	// 4. Apply Filters
	validDrivers := candidateDrivers
	for _, filter := range s.filters {
		validDrivers = filter.Filter(validDrivers, booking)
	}

	if len(validDrivers) == 0 {
		log.Info().Str("booking_id", bookingID.String()).Msg("No matching drivers after filtering")
		return
	}

	if err := s.bookingRepo.AddNotifiedDrivers(ctx, bookingID, validDrivers); err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Error saving notified drivers")
		return
	}

	// 5. Notify
	log.Info().
		Str("booking_id", bookingID.String()).
		Int("driver_count", len(candidateDrivers)).
		Msg("Found matching drivers. Notifying...")

	for _, d := range validDrivers {
		// Mock Notification
		// TODO: In a real app, this would be an async call to a notification service
		// TODO: Integrate with real notification service (e.g., Firebase, Twilio)
		// For now, we just log the notification
		log.Info().
			Str("booking_id", bookingID.String()).
			Str("driver_id", d.ID.String()).
			Str("driver_name", d.Name).
			Str("phone", d.PhoneNumber).
			Msg(">> Push Notification Sent")
	}
}
