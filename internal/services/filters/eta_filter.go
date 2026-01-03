package filters

import (
	"CabBookingService/internal/models"
	"CabBookingService/internal/util"
)

type etaBasedFilter struct {
	maxDistanceKm float64
}

func NewETABasedFilter(maxDistanceKm float64) DriverFilter {
	return &etaBasedFilter{
		maxDistanceKm: maxDistanceKm,
	}
}
func (f *etaBasedFilter) Filter(drivers []models.Driver, booking *models.Booking) []models.Driver {
	validDrivers := make([]models.Driver, 0)

	// NOTE: In a real ETA filter, you would call Google Maps API here.
	// For this implementation, we will use the "Naive" distance calculation
	// we already have, filtering out drivers who might have drifted too far
	// since the initial spatial search.

	for _, driver := range drivers {
		// We assume driver.LastKnownLocation is populated
		if driver.LastKnownLocation == nil {
			continue
		}

		// Use the Haversine formula (exported from your location service or a util package)
		distance := util.DistanceKm(
			booking.PickupLatitude, booking.PickupLongitude,
			driver.LastKnownLocation.Latitude, driver.LastKnownLocation.Longitude,
		)

		if distance <= f.maxDistanceKm {
			validDrivers = append(validDrivers, driver)
		}
	}
	return validDrivers
}
