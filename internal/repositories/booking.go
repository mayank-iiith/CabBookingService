package repositories

import (
	"CabBookingService/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookingRepository defines the methods for interacting with booking data.
// By using an interface, our services can be tested with mocks
// and we can easily swap GORM for another DB if needed.
type BookingRepository interface {
	Create(booking *models.Booking) error
	GetByID(id uuid.UUID) (*models.Booking, error)
	Update(booking *models.Booking) error
	UpdateStatus(id uuid.UUID, status models.BookingStatus) error

	// New Security Methods
	AddNotifiedDrivers(bookingID uuid.UUID, drivers []models.Driver) error
	IsDriverNotified(bookingID uuid.UUID, driverID uuid.UUID) (bool, error)

	SaveReviewAndRecalculateDriverRating(bookingID uuid.UUID, review *models.Review) error
	SaveReviewAndRecalculatePassengerRating(bookingID uuid.UUID, review *models.Review) error

	GetPendingBookingsForDriver(driverID uuid.UUID) ([]models.Booking, error)
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

func (r *gormBookingRepository) Create(booking *models.Booking) error {
	if err := r.db.Create(booking).Error; err != nil {
		return err
	}
	return nil
}

func (r *gormBookingRepository) GetByID(id uuid.UUID) (*models.Booking, error) {
	var booking models.Booking
	if err := r.db.
		Preload("Passenger").
		Preload("Driver").
		Preload("RideStartOTP").
		First(&booking, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *gormBookingRepository) Update(booking *models.Booking) error {
	return r.db.Save(booking).Error
}

func (r *gormBookingRepository) UpdateStatus(id uuid.UUID, status models.BookingStatus) error {
	return r.db.Model(&models.Booking{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *gormBookingRepository) AddNotifiedDrivers(bookingID uuid.UUID, drivers []models.Driver) error {
	booking := models.Booking{BaseModel: models.BaseModel{ID: bookingID}}
	// GORM's Association Mode handles the INSERT into booking_notified_drivers
	return r.db.Model(&booking).Association("NotifiedDrivers").Replace(drivers)
}

func (r *gormBookingRepository) IsDriverNotified(bookingID uuid.UUID, driverID uuid.UUID) (bool, error) {
	var count int64
	// Raw SQL check is often faster/simpler for existence checks
	err := r.db.Table("booking_notified_drivers").
		Where("booking_id = ? AND driver_id = ?", bookingID, driverID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *gormBookingRepository) SaveReviewAndRecalculateDriverRating(bookingID uuid.UUID, review *models.Review) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

		if err := tx.Model(&models.Driver{}).Where("id = ?", review.DriverID).Updates(map[string]interface{}{
			"average_rating": newAvg,
			"rating_count":   newCount,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormBookingRepository) SaveReviewAndRecalculatePassengerRating(bookingID uuid.UUID, review *models.Review) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *gormBookingRepository) GetPendingBookingsForDriver(driverID uuid.UUID) ([]models.Booking, error) {
	var bookings []models.Booking

	err := r.db.
		Table("bookings").
		Joins("JOIN booking_notified_drivers ON bookings.id = booking_notified_drivers.booking_id").
		Where("booking_notified_drivers.driver_id = ?", driverID).
		Where("bookings.status = ?", models.BookingStatusRequested).
		Preload("Passenger").
		Preload("Passenger.Account").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	fmt.Println("bookings:", bookings)

	return bookings, nil
}
