package models

type VehicleLocation struct {
	ID        int64   `json:"id,omitempty"`
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type GeofenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"` // "geofence_entry"

	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`

	Timestamp int64 `json:"timestamp"`
}
