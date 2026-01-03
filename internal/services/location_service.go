package services

import (
	"context"
	"sync"

	"CabBookingService/internal/repositories"
	"CabBookingService/internal/util"

	"github.com/google/uuid"
)

// driverLocation is a simple struct to hold coordinates
type driverLocation struct {
	latitude  float64
	longitude float64
}

type LocationService interface {
	UpdateDriverLocation(ctx context.Context, driverID uuid.UUID, latitude, longitude float64) error
	GetNearbyDrivers(lat, lon float64, radiusKm float64) []uuid.UUID
}

// NaiveLocationService uses a map and loops through all drivers.
// TODO: Replace this with a QuadTree for production (O(N) vs O(log N))
type naiveLocationService struct {
	driverRepo      repositories.DriverRepository
	driverLocations map[uuid.UUID]driverLocation
	mu              sync.RWMutex
}

func NewNaiveLocationService(driverRepo repositories.DriverRepository) LocationService {
	return &naiveLocationService{
		driverRepo:      driverRepo,
		driverLocations: make(map[uuid.UUID]driverLocation),
	}
}

func (s *naiveLocationService) UpdateDriverLocation(ctx context.Context, driverID uuid.UUID, latitude, longitude float64) error {

	// 1. Update in memory naive driver location map (fast)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.driverLocations[driverID] = driverLocation{
		latitude:  latitude,
		longitude: longitude,
	}

	// 2. Persist to DB (Reliable)
	// We do this async or strictly depending on requirements.
	// For now simply do it synchronously.
	return s.driverRepo.UpdateLocation(ctx, driverID, latitude, longitude)
}

func (s *naiveLocationService) GetNearbyDrivers(lat, lon float64, radiusKm float64) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nearbyDrivers []uuid.UUID

	// Naive Approach: Check every single driver
	for id, location := range s.driverLocations {
		if util.DistanceKm(lat, lon, location.latitude, location.longitude) <= radiusKm {
			nearbyDrivers = append(nearbyDrivers, id)
		}
	}
	return nearbyDrivers
}
