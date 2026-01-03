package repositories

import (
	"CabBookingService/internal/db"
	"CabBookingService/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookingRepository defines the methods for interacting with booking data.
// By using an interface, our services can be tested with mocks
// and we can easily swap GORM for another DB if needed.
type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error)
	Update(ctx context.Context, booking *models.Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.BookingStatus) error

	// New Security Methods
	AddNotifiedDrivers(ctx context.Context, bookingID uuid.UUID, drivers []models.Driver) error
	IsDriverNotified(ctx context.Context, bookingID uuid.UUID, driverID uuid.UUID) (bool, error)

	SaveReviewAndRecalculateDriverRating(ctx context.Context, bookingID uuid.UUID, review *models.Review) error
	SaveReviewAndRecalculatePassengerRating(ctx context.Context, bookingID uuid.UUID, review *models.Review) error

	GetPendingBookingsForDriver(ctx context.Context, driverID uuid.UUID, limit, offset int) ([]models.Booking, error)

	AssignDriverIfAvailable(ctx context.Context, bookingID uuid.UUID, driverID uuid.UUID) error

	GetDueScheduledBookings(ctx context.Context, cutoff time.Time) ([]models.Booking, error)
}

type gormBookingRepository struct {
	db *gorm.DB // The GORM database connection
}

// Ensure gormBookingRepository implements BookingRepository
var _ BookingRepository = &gormBookingRepository{nil}

// NewGormBookingRepository creates a new GORM booking repository
func NewGormBookingRepository(db *gorm.DB) BookingRepository {
	return &gormBookingRepository{db: db}
}

func (r *gormBookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(booking).Error
}

func (r *gormBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	tx := db.NewGormTx(ctx, r.db)

	var booking models.Booking
	err := tx.Model(&models.Booking{}).
		Preload("Passenger").
		Preload("Driver").
		Preload("RideStartOTP").
		Preload("ReviewByPassenger").
		Preload("ReviewByDriver").
		First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *gormBookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Save(booking).Error
}

func (r *gormBookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.BookingStatus) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Model(&models.Booking{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *gormBookingRepository) AddNotifiedDrivers(ctx context.Context, bookingID uuid.UUID, drivers []models.Driver) error {
	tx := db.NewGormTx(ctx, r.db)
	booking := models.Booking{BaseModel: models.BaseModel{ID: bookingID}}
	// GORM's Association Mode handles the INSERT into booking_notified_drivers
	return tx.Model(&booking).Association("NotifiedDrivers").Replace(drivers)
}

func (r *gormBookingRepository) IsDriverNotified(ctx context.Context, bookingID uuid.UUID, driverID uuid.UUID) (bool, error) {
	tx := db.NewGormTx(ctx, r.db)
	var count int64
	err := tx.Table("booking_notified_drivers").
		Where("booking_id = ? AND driver_id = ?", bookingID, driverID).
		Count(&count).Error
	return count > 0, err
}

func (r *gormBookingRepository) SaveReviewAndRecalculateDriverRating(ctx context.Context, bookingID uuid.UUID, review *models.Review) error {
	tx := db.NewGormTx(ctx, r.db)

	return tx.Transaction(func(tx *gorm.DB) error {
		// 1. Save the Review
		if err := tx.Create(review).Error; err != nil {
			return err
		}

		// 2. Link Review to Booking
		if err := tx.Model(&models.Booking{}).Where("id = ?", bookingID).
			Update("review_by_passenger_id", review.ID).Error; err != nil {
			return err
		}

		// 3. Recalculate Driver's Average Rating
		// NewAvg = ((OldAvg * Count) + NewRating) / (Count + 1)
		var driver models.Driver
		if err := tx.First(&driver, "id = ?", review.DriverID).Error; err != nil {
			return err
		}

		newCount := driver.RatingCount + 1
		newAvg := ((driver.AverageRating * float64(driver.RatingCount)) + float64(review.Rating)) / float64(newCount)
		// TODO: round newAvg to 2 decimal places if needed

		// 4. Update Driver's Average Rating and Rating Count
		if err := tx.Model(&models.Driver{}).Where("id = ?", review.DriverID).Updates(map[string]interface{}{
			"average_rating": newAvg,
			"rating_count":   newCount,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormBookingRepository) SaveReviewAndRecalculatePassengerRating(ctx context.Context, bookingID uuid.UUID, review *models.Review) error {
	tx := db.NewGormTx(ctx, r.db)

	return tx.Transaction(func(tx *gorm.DB) error {
		// 1. Save the Review
		if err := tx.Create(review).Error; err != nil {
			return err
		}

		// 2. Link Review to Booking
		if err := tx.Model(&models.Booking{}).Where("id = ?", bookingID).
			Update("review_by_passenger_id", review.ID).Error; err != nil {
			return err
		}

		// 3. Recalculate Passenger's Average Rating
		// NewAvg = ((OldAvg * Count) + NewRating) / (Count + 1)
		var passenger models.Passenger
		if err := tx.First(&passenger, "id = ?", review.DriverID).Error; err != nil {
			return err
		}

		newCount := passenger.RatingCount + 1
		newAvg := ((passenger.AverageRating * float64(passenger.RatingCount)) + float64(review.Rating)) / float64(newCount)

		if err := tx.Model(&models.Passenger{}).Where("id = ?", review.PassengerId).Updates(map[string]interface{}{
			"average_rating": newAvg,
			"rating_count":   newCount,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormBookingRepository) GetPendingBookingsForDriver(ctx context.Context, driverID uuid.UUID, limit, offset int) ([]models.Booking, error) {
	tx := db.NewGormTx(ctx, r.db)

	var bookings []models.Booking
	err := tx.Table("bookings").
		Joins("JOIN booking_notified_drivers ON bookings.id = booking_notified_drivers.booking_id").
		Where("booking_notified_drivers.driver_id = ?", driverID).
		Where("bookings.status = ?", models.BookingStatusRequested).
		Limit(limit).
		Offset(offset).
		Preload("Passenger").
		Preload("Passenger.Account").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

// Implementation
func (r *gormBookingRepository) AssignDriverIfAvailable(ctx context.Context, bookingID uuid.UUID, driverID uuid.UUID) error {
	tx := db.NewGormTx(ctx, r.db)

	// UPDATE bookings SET status='ACCEPTED', driver_id=?
	// WHERE id=? AND status='REQUESTED'
	result := tx.Model(&models.Booking{}).
		Where("id = ? AND status = ?", bookingID, models.BookingStatusRequested).
		Updates(map[string]interface{}{
			"status":    models.BookingStatusAccepted,
			"driver_id": driverID,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Or custom error: "booking not available"
	}

	return nil
}

// GetDueScheduledBookings finds bookings that are in SCHEDULED status and ready to be processed
func (r *gormBookingRepository) GetDueScheduledBookings(ctx context.Context, cutoff time.Time) ([]models.Booking, error) {
	tx := db.NewGormTx(ctx, r.db)

	var bookings []models.Booking
	// We check for:
	// 1. Status is 'SCHEDULED'
	// 2. ScheduledTime is before or equal to the cutoff (e.g., Now + 15 mins)
	// TODO: Take care of pagination if needed
	err := tx.Model(&models.Booking{}).
		Where("status = ? AND scheduled_time <= ?", models.BookingStatusScheduled, cutoff).
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	return nil, nil
}
