package geofence

import (
	"math"
)

type Geofence struct {
	Lat     float64
	Lon     float64
	RadiusM float64
}

// NewGeofence membuat geofence lingkaran sederhana.
func NewGeofence(lat, lon, radius float64) *Geofence {
	return &Geofence{
		Lat:     lat,
		Lon:     lon,
		RadiusM: radius,
	}
}

// Haversine distance (meter) antara 2 koordinat.
func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // radius bumi dalam meter
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// IsInside mengecek apakah titik berada di dalam radius geofence.
func (g *Geofence) IsInside(lat, lon float64) bool {
	return distanceMeters(g.Lat, g.Lon, lat, lon) <= g.RadiusM
}
