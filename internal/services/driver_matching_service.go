package services

import (
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services/queue"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

const (
	TopicDriverMatching = "DRIVER_MATCHING"
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
	}
}

func (s *driverMatchingService) StartConsuming() error {
	ch, err := s.queue.Subscribe(TopicDriverMatching)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", TopicDriverMatching, err)
	}

	// Start a background goroutine to consume messages
	go func() {
		log.Println("[DriverMatching] Started consuming messages...")
		for msg := range ch {
			// Type Assertion
			bookingID, ok := msg.(uuid.UUID)
			if !ok {
				log.Printf("[DriverMatching] Invalid message format received, msg: %v\n", msg)
				continue
				// TODO: In production, consider dead-letter queue or alerting
			}
			s.handleDriverMatching(bookingID)
		}
	}()
	return nil
}

func (s *driverMatchingService) handleDriverMatching(bookingID uuid.UUID) {
	log.Printf("[DriverMatching] Handling driver matching for booking ID: %s\n", bookingID)
	ctx := context.Background() // TODO: Either pass context from message or create with timeout

	// 1. Fetch Booking
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		log.Printf("Error fetching booking %s: %v", bookingID, err)
		return
	}

	// 2. Find nearby drivers using pickup location from booking
	radiusToSearch := 2.0 // in km
	nearbyDriverIDs := s.locationService.GetNearbyDrivers(booking.PickupLatitude, booking.PickupLongitude, radiusToSearch)
	if len(nearbyDriverIDs) == 0 {
		log.Printf("No drivers nearby for booking %s", bookingID)
		return
	}

	// 3. Fetch Full Driver Profiles
	candidateDrivers, err := s.driverRepo.GetByAccountIDs(ctx, nearbyDriverIDs)
	if err != nil {
		log.Printf("Error fetching driver profiles for booking %s: %v", bookingID, err)
		return
	}

	// 4. Apply Filters
	validDrivers := candidateDrivers
	//for _, filter := range s.filters {
	//	validDrivers = filter.Filter(validDrivers, *booking)
	//}

	if len(validDrivers) == 0 {
		log.Printf("No matching drivers after filtering for booking %s", bookingID)
		return
	}

	if err := s.bookingRepo.AddNotifiedDrivers(ctx, bookingID, validDrivers); err != nil {
		log.Printf("Error adding notified drivers for booking %s: %v", bookingID, err)
		return
	}

	// 5. Notify
	log.Printf("Found %d matching drivers for booking %s. Notifying...", len(validDrivers), bookingID)
	for _, d := range validDrivers {
		// Mock Notification
		log.Printf(">> Sending Push Notification to Driver: %s (Phone: %s)", d.Name, d.PhoneNumber)
		// TODO: Integrate with real notification service (e.g., Firebase, Twilio)
		// For now, we just log the notification
	}
}
