package util

import (
	"math"
)

// DistanceKm calculates the distance between two points using the Haversine formula
// (Or a simplified version for small distances)
// TODO: Learn about Haversine formula and implement it properly
func DistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
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
