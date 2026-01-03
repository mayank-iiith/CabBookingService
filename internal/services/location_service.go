package services

import (
	"math"
	"sync"

	"github.com/google/uuid"
)

// driverLocation is a simple struct to hold coordinates
type driverLocation struct {
	latitude  float64
	longitude float64
}

type LocationService interface {
	UpdateDriverLocation(driverID uuid.UUID, latitude, longitude float64) error
	GetNearbyDrivers(lat, lon float64, radiusKm float64) []uuid.UUID
}

// NaiveLocationService uses a map and loops through all drivers.
// TODO: Replace this with a QuadTree for production (O(N) vs O(log N))
type naiveLocationService struct {
	driverLocations map[uuid.UUID]driverLocation
	mu              sync.RWMutex
}

func NewNaiveLocationService() LocationService {
	return &naiveLocationService{
		driverLocations: make(map[uuid.UUID]driverLocation),
	}
}

func (s *naiveLocationService) UpdateDriverLocation(driverID uuid.UUID, latitude, longitude float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.driverLocations[driverID] = driverLocation{
		latitude:  latitude,
		longitude: longitude,
	}

	return nil
}

func (s *naiveLocationService) GetNearbyDrivers(lat, lon float64, radiusKm float64) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nearbyDrivers []uuid.UUID

	// Naive Approach: Check every single driver
	for id, location := range s.driverLocations {
		if distanceKm(lat, lon, location.latitude, location.longitude) <= radiusKm {
			nearbyDrivers = append(nearbyDrivers, id)
		}
	}
	return nearbyDrivers
}

// distanceKm calculates the distance between two points using the Haversine formula
// (Or a simplified version for small distances)
// TODO: Learn about Haversine formula and implement it properly
func distanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	// Radius of the Earth in km
	const R = 6371.0

	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
