package services

import (
	"context"
	"errors"
	"time"

	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services/queue"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type BookingService interface {
	CreateBooking(ctx context.Context, passengerAccountID uuid.UUID, pickupLatitude, pickupLongitude, dropoffLatitude, dropoffLongitude float64, scheduledTime *time.Time) (*models.Booking, error)
	AcceptBooking(ctx context.Context, driverAccountID, bookingID uuid.UUID) error
	CancelBooking(ctx context.Context, driverAccountID, bookingID uuid.UUID) error
	StartRide(ctx context.Context, driverAccountID, bookingID uuid.UUID, otpCode string) error
	EndRide(ctx context.Context, driverAccountID, bookingID uuid.UUID) error
	RateRide(ctx context.Context, bookingID uuid.UUID, rating int, note string, isPassenger bool) error
	GetPendingRides(ctx context.Context, driverAccountID uuid.UUID, limit, offset int) ([]models.Booking, error)

	// TODO: Move to DriverService?
	ToggleDriverAvailability(ctx context.Context, driverAccountID uuid.UUID, available bool) error

	// TODO: Add method AssignDriver(bookingID, driverID)
}

type bookingService struct {
	bookingRepo     repositories.BookingRepository
	driverRepo      repositories.DriverRepository
	passengerRepo   repositories.PassengerRepository
	reviewRepo      repositories.ReviewRepository
	otpService      OTPService
	locationService LocationService
	paymentService  PaymentService
	messageQueue    queue.MessageQueue
}

func NewBookingService(
	bookingRepo repositories.BookingRepository,
	driverRepo repositories.DriverRepository,
	passengerRepo repositories.PassengerRepository,
	reviewRepo repositories.ReviewRepository,
	otpService OTPService,
	locationService LocationService,
	paymentService PaymentService,
	messageQueue queue.MessageQueue,
) BookingService {
	return &bookingService{
		bookingRepo:     bookingRepo,
		driverRepo:      driverRepo,
		passengerRepo:   passengerRepo,
		reviewRepo:      reviewRepo,
		otpService:      otpService,
		locationService: locationService,
		paymentService:  paymentService,
		messageQueue:    messageQueue,
	}
}

// CreateBooking Passenger requests a ride
func (b *bookingService) CreateBooking(ctx context.Context, passengerAccountID uuid.UUID, pickupLatitude, pickupLongitude, dropoffLatitude, dropoffLongitude float64, scheduledTime *time.Time) (*models.Booking, error) {
	// 1. Get Passenger Profile from Account ID
	passenger, err := b.passengerRepo.GetByAccountID(ctx, passengerAccountID)
	if err != nil {
		return nil, err
	}

	// 2. Generate OTP for ride start
	otp, err := b.otpService.GenerateOTP(ctx, passenger.PhoneNumber)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	status := models.BookingStatusRequested
	// If scheduled time is > 20 mins from now, set as SCHEDULED
	if scheduledTime != nil && scheduledTime.After(now.Add(20*time.Minute)) {
		status = models.BookingStatusScheduled
	}

	// 3. Create Booking
	booking := &models.Booking{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		PassengerId:    passenger.ID,
		Status:         status,
		RideStartOTPId: &otp.ID,

		PickupLatitude:   pickupLatitude,
		PickupLongitude:  pickupLongitude,
		DropoffLatitude:  dropoffLatitude,
		DropoffLongitude: dropoffLongitude,
		ScheduledTime:    scheduledTime,
	}

	if err := b.bookingRepo.Create(ctx, booking); err != nil {
		log.Error().Err(err).Msg("Failed to create booking record")
		return nil, err
	}

	// Structured Log for traceability
	log.Info().
		Str("booking_id", booking.ID.String()).
		Str("passenger_id", passenger.ID.String()).
		Msg("Booking created successfully")

	// --- ASYNC LOGIC ---
	// Instead of calling location service directly, we push to queue
	// This makes the API response fast (Fire and Forget)
	// Only push to queue if status is REQUESTED
	if booking.Status == models.BookingStatusRequested {
		if err := b.messageQueue.Publish(ctx, TopicDriverMatching, booking.ID); err != nil {
			// Log error but don't fail request
			log.Error().Err(err).
				Str("booking_id", booking.ID.String()).
				Msg("Failed to publish driver matching event")
			// Don't fail the request, just log it (or implement retry)
		}
	}

	// For scheduled rides, we rely on SchedulingService to pick them up later
	// when their scheduled time is near.
	// So no need to push to queue now.

	return booking, nil
}

// AcceptBooking Driver accepts a ride
func (b *bookingService) AcceptBooking(ctx context.Context, driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return err
	}

	// 2. Check Permissions (Security) - Check if this driver was actually notified/authorized for this ride
	allowed, err := b.bookingRepo.IsDriverNotified(ctx, bookingID, driver.ID)
	if err != nil {
		return errors.New("system error checking permissions")
	}
	if !allowed {
		log.Warn().
			Str("driver_id", driver.ID.String()).
			Str("booking_id", bookingID.String()).
			Msg("Driver attempted to accept booking without the ride assignment to them")
		return errors.New("you are not authorized to accept this ride")
	}

	// 3. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// 4. Get Passenger to generate OTP
	passenger, err := b.passengerRepo.GetByID(ctx, booking.PassengerId) // Need to fetch passenger to get phone
	if err != nil {
		return err
	}

	// 5. Generate OTP now for Ride Start as the driver is assigned
	otp, err := b.otpService.GenerateOTP(ctx, passenger.PhoneNumber)
	if err != nil {
		return err
	}

	// 3. Accept Booking Transactionally (accept and assign driver, mark driver unavailable)
	if err := b.bookingRepo.AcceptBookingTransaction(ctx, bookingID, driver.ID, otp.ID); err != nil {
		return err
	}
	log.Info().
		Str("booking_id", bookingID.String()).
		Str("driver_id", driver.ID.String()).
		Msg("Booking accepted by driver")

	return nil
}

// CancelBooking Driver cancels a ride
func (b *bookingService) CancelBooking(ctx context.Context, driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// 3. Authorization Check (Check if Driver is assigned to this Booking)
	if booking.DriverId == nil || *booking.DriverId != driver.ID {
		return errors.New("driver not assigned to this booking")
	}

	if booking.Status.IsCancellable() {
		return errors.New("booking cannot be cancelled at this stage")
	}

	// 4. Update Booking Status to REQUESTED and remove Driver assignment
	booking.Status = models.BookingStatusCancelled
	booking.DriverId = nil

	if err := b.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	log.Info().
		Str("booking_id", booking.ID.String()).
		Str("driver_id", driver.ID.String()).
		Msg("Booking cancelled by driver")

	// TODO: Notify Passenger about cancellation
	// TODO: Re-emit event to "DriverMatchingService" to find another driver

	return b.driverRepo.UpdateAvailability(ctx, driver.ID, true)
}

// StartRide Driver verifies OTP and starts
func (b *bookingService) StartRide(ctx context.Context, driverAccountID, bookingID uuid.UUID, otpCode string) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// 3. Authorization Check (Check if Driver is assigned to this Booking)
	if booking.DriverId == nil || *booking.DriverId != driver.ID {
		return errors.New("driver not assigned to this booking")
	}

	if booking.Status != models.BookingStatusAccepted {
		return errors.New("booking is not in accepted status")
	}

	// 4. Validate OTP
	if booking.RideStartOTPId == nil || !b.otpService.ValidateOTP(ctx, *booking.RideStartOTPId, otpCode) {
		log.Warn().
			Str("booking_id", booking.ID.String()).
			Str("driver_id", driver.ID.String()).
			Msg("Invalid OTP provided to start ride")
		return errors.New("invalid OTP code")
	}

	// 5. Update Booking Status to STARTED
	booking.Status = models.BookingStatusStarted
	if err := b.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}
	log.Info().Str("booking_id", bookingID.String()).Msg("Ride started")
	return nil
}

// EndRide Driver completes the trip
func (b *bookingService) EndRide(ctx context.Context, driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// 3. Authorization Check (Check if Driver is assigned to this Booking)
	if booking.DriverId == nil || *booking.DriverId != driver.ID {
		return errors.New("driver not assigned to this booking")
	}

	if booking.Status != models.BookingStatusStarted {
		return errors.New("booking is not in started status")
	}

	// 4. Update Booking Status to COMPLETED
	booking.Status = models.BookingStatusCompleted
	if err := b.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	log.Info().
		Str("booking_id", bookingID.String()).
		Str("driver_id", driver.ID.String()).
		Msg("Ride completed by driver")

	// 3. --- TRIGGER PAYMENT ---
	// This happens asynchronously in real life, but sync here for simplicity
	if err := b.paymentService.ProcessPayment(ctx, booking); err != nil {
		// Log error, but don't fail the ride completion.
		// In real world, this would trigger a "Payment Failed" flow.
		log.Error().Err(err).Msg("Payment processing failed")
	}

	return b.driverRepo.UpdateAvailability(ctx, driver.ID, true)
}

func (b *bookingService) RateRide(ctx context.Context, bookingID uuid.UUID, rating int, note string, isPassenger bool) error {
	// 1. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != models.BookingStatusCompleted {
		return errors.New("ride can be rated only after completion")
	}

	// 2. Create Review
	review := &models.Review{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Rating:    rating,
		Note:      note,
		BookingId: booking.ID,
	}

	if isPassenger {
		review.PassengerId = &booking.PassengerId
		return b.bookingRepo.SaveReviewAndRecalculateDriverRating(ctx, bookingID, review)
	}

	if booking.DriverId == nil {
		// Ideally this should not happen if the ride was completed
		log.Warn().
			Str("booking_id", booking.ID.String()).
			Msg("No driver assigned to booking while rating by driver")
		return errors.New("no driver assigned")
	}
	review.DriverID = booking.DriverId
	return b.bookingRepo.SaveReviewAndRecalculatePassengerRating(ctx, bookingID, review)
}

func (b *bookingService) GetPendingRides(ctx context.Context, driverAccountID uuid.UUID, limit, offset int) ([]models.Booking, error) {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch Pending Rides
	return b.bookingRepo.GetPendingBookingsForDriver(ctx, driver.ID, limit, offset)
}

func (b *bookingService) ToggleDriverAvailability(ctx context.Context, driverAccountID uuid.UUID, available bool) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(ctx, driverAccountID)
	if err != nil {
		return err
	}

	// 2. Update Availability
	return b.driverRepo.UpdateAvailability(ctx, driver.ID, available)
}
