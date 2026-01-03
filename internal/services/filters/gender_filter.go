package filters

import (
	"CabBookingService/internal/models"
)

type genderFilter struct{}

func NewGenderFilter() DriverFilter {
	return &genderFilter{}
}

func (*genderFilter) Filter(drivers []models.Driver, booking *models.Booking) []models.Driver {

	// Passenger Gender - Fallback if gender is nil
	passengerGender := models.GenderOther
	if booking.Passenger.Gender != nil {
		passengerGender = *booking.Passenger.Gender
	}

	validDrivers := make([]models.Driver, 0)
	for _, driver := range drivers {
		driverGender := models.GenderOther
		if driver.Gender != nil {
			driverGender = *driver.Gender
		}

		// "Male drivers can only drive male passengers" (Safety feature in this specific domain logic)
		// If Driver is Male, Passenger MUST be Male.
		// If Driver is Female/Other, they can drive anyone.
		// TODO: Add more robust logic based on real-world requirements.

		// If driver is Male AND Passenger is NOT Male -> Skip
		if driverGender == models.GenderMale && passengerGender != models.GenderMale {
			continue
		}
		validDrivers = append(validDrivers, driver)
	}
	return validDrivers
}
