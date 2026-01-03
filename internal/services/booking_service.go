package services

import (
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services/queue"
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(passengerAccountID uuid.UUID, pickupLatitude, pickupLongitude, dropoffLatitude, dropoffLongitude float64) (*models.Booking, error)
	AcceptBooking(driverAccountID, bookingID uuid.UUID) error
	CancelBooking(driverAccountID, bookingID uuid.UUID) error
	StartRide(driverAccountID, bookingID uuid.UUID, otpCode string) error
	EndRide(driverAccountID, bookingID uuid.UUID) error
	RateRide(bookingID uuid.UUID, rating int, note string, isPassenger bool) error
	GetPendingRides(driverAccountID uuid.UUID) ([]models.Booking, error)

	// TODO: Add method AssignDriver(bookingID, driverID)
}

type bookingService struct {
	bookingRepo     repositories.BookingRepository
	driverRepo      repositories.DriverRepository
	passengerRepo   repositories.PassengerRepository
	reviewRepo      repositories.ReviewRepository
	otpService      OTPService
	locationService LocationService
	messageQueue    queue.MessageQueue
}

func NewBookingService(
	bookingRepo repositories.BookingRepository,
	driverRepo repositories.DriverRepository,
	passengerRepo repositories.PassengerRepository,
	reviewRepo repositories.ReviewRepository,
	otpService OTPService,
	locationService LocationService,
	messageQueue queue.MessageQueue,
) BookingService {
	return &bookingService{
		bookingRepo:     bookingRepo,
		driverRepo:      driverRepo,
		passengerRepo:   passengerRepo,
		reviewRepo:      reviewRepo,
		otpService:      otpService,
		locationService: locationService,
		messageQueue:    messageQueue,
	}
}

// CreateBooking Passenger requests a ride
func (b *bookingService) CreateBooking(passengerAccountID uuid.UUID, pickupLatitude, pickupLongitude, dropoffLatitude, dropoffLongitude float64) (*models.Booking, error) {
	// 1. Get Passenger Profile from Account ID
	passenger, err := b.passengerRepo.GetByAccountID(passengerAccountID)
	if err != nil {
		return nil, err
	}
	slog.Debug("[BookingService] CreateBooking", slog.String("passengerID", passenger.ID.String()), slog.String("accountID", passengerAccountID.String()))

	// 2. Generate OTP for ride start
	otp, err := b.otpService.GenerateOTP(passenger.PhoneNumber)
	if err != nil {
		return nil, err
	}
	slog.Debug("[BookingService] Generated OTP", slog.String("otpID", otp.ID.String()), slog.String("passengerID", passenger.ID.String()), slog.String("accountID", passengerAccountID.String()), slog.String("otpCode", otp.Code))

	now := time.Now()

	// 3. Create Booking
	booking := &models.Booking{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		PassengerId:    passenger.ID,
		Status:         models.BookingStatusRequested,
		RideStartOTPId: &otp.ID,

		PickupLatitude:   pickupLatitude,
		PickupLongitude:  pickupLongitude,
		DropoffLatitude:  dropoffLatitude,
		DropoffLongitude: dropoffLongitude,
	}

	if err := b.bookingRepo.Create(booking); err != nil {
		return nil, err
	}
	slog.Debug("[BookingService] Created Booking", slog.String("bookingID", booking.ID.String()), slog.String("passengerID", passenger.ID.String()), slog.String("accountID", passengerAccountID.String()))

	// --- ASYNC LOGIC ---
	// Instead of calling location service directly, we push to queue
	// This makes the API response fast (Fire and Forget)
	if err := b.messageQueue.Publish(TopicDriverMatching, booking.ID); err != nil {
		log.Printf("Failed to publish driver matching event: %v", err)
		// Don't fail the request, just log it (or implement retry)
	}

	booking.RideStartOTP = otp
	return booking, nil
}

// AcceptBooking Driver accepts a ride
func (b *bookingService) AcceptBooking(driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	// 3. Check if Booking is still in REQUESTED status
	if booking.Status != models.BookingStatusRequested {
		//return ErrBookingNotAvailable
		return errors.New("booking is no longer available")
	}

	// Check if this driver was actually notified/authorized for this ride
	allowed, err := b.bookingRepo.IsDriverNotified(bookingID, driver.ID)
	if err != nil {
		return errors.New("system error checking permissions")
	}
	if !allowed {
		return errors.New("you are not authorized to accept this ride")
	}

	// TODO: Need to make sure two drivers don't accept the same ride at the same time
	// This can be handled via DB transactions or optimistic locking
	// Also as soon as a driver accepts, we need to stop displaying this booking to other drivers

	// 4. Assign Driver
	booking.DriverId = &driver.ID
	booking.Status = models.BookingStatusAccepted

	return b.bookingRepo.Update(booking)
}

// CancelBooking Driver cancels a ride
func (b *bookingService) CancelBooking(driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	// 3. Authorization Check (Check if Driver is assigned to this Booking)
	if booking.DriverId == nil || *booking.DriverId != driver.ID {
		return errors.New("driver not assigned to this booking")
	}

	if booking.Status != models.BookingStatusRequested && booking.Status != models.BookingStatusAccepted {
		return errors.New("booking cannot be cancelled at this stage")
	}

	// 4. Update Booking Status to REQUESTED and remove Driver assignment
	booking.Status = models.BookingStatusCancelled
	booking.DriverId = nil

	// TODO: Notify Passenger about cancellation
	// TODO: Re-emit event to "DriverMatchingService" to find another driver

	return b.bookingRepo.Update(booking)
}

// StartRide Driver verifies OTP and starts
func (b *bookingService) StartRide(driverAccountID, bookingID uuid.UUID, otpCode string) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(bookingID)
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
	if booking.RideStartOTPId == nil || !b.otpService.ValidateOTP(*booking.RideStartOTPId, otpCode) {
		return errors.New("invalid OTP code")
	}

	// 5. Update Booking Status to STARTED
	booking.Status = models.BookingStatusStarted
	return b.bookingRepo.Update(booking)
}

// EndRide Driver completes the trip
func (b *bookingService) EndRide(driverAccountID, bookingID uuid.UUID) error {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(driverAccountID)
	if err != nil {
		return err
	}

	// 2. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(bookingID)
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
	return b.bookingRepo.Update(booking)
}

func (b *bookingService) RateRide(bookingID uuid.UUID, rating int, note string, isPassenger bool) error {
	// 1. Get Booking by ID
	booking, err := b.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
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
		return b.bookingRepo.SaveReviewAndRecalculateDriverRating(bookingID, review)
	}

	if booking.DriverId == nil {
		return errors.New("no driver assigned to this booking")
	}
	review.DriverID = booking.DriverId
	return b.bookingRepo.SaveReviewAndRecalculatePassengerRating(bookingID, review)
}

func (b *bookingService) GetPendingRides(driverAccountID uuid.UUID) ([]models.Booking, error) {
	// 1. Get Driver Profile from Account ID
	driver, err := b.driverRepo.GetByAccountID(driverAccountID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch Pending Rides
	return b.bookingRepo.GetPendingBookingsForDriver(driver.ID)
}
