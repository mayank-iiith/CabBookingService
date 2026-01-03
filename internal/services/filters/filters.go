package filters

import (
	"CabBookingService/internal/models"
)

// DriverFilter defines the contract for filtering drivers based on booking criteria
type DriverFilter interface {
	Filter(drivers []models.Driver, booking *models.Booking) []models.Driver
}
