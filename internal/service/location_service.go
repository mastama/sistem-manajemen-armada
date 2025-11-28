package service

import (
	"context"
	"errors"
	"log"

	"sistem-manajemen-armada/internal/geofence"
	"sistem-manajemen-armada/internal/models"
	"sistem-manajemen-armada/internal/rabbitmq"
	"sistem-manajemen-armada/internal/repository"
)

type LocationService struct {
	repo      repository.LocationRepository
	geofence  *geofence.Geofence
	rabbitCli *rabbitmq.Client
}

func NewLocationService(repo repository.LocationRepository, g *geofence.Geofence, r *rabbitmq.Client) *LocationService {
	return &LocationService{
		repo:      repo,
		geofence:  g,
		rabbitCli: r,
	}
}

func (s *LocationService) SaveLocation(ctx context.Context, loc models.VehicleLocation) error {
	if loc.VehicleID == "" {
		return errors.New("vehicle_id is required")
	}
	if loc.Timestamp == 0 {
		return errors.New("timestamp is required")
	}

	if err := s.repo.Insert(ctx, loc); err != nil {
		return err
	}

	// Cek geofence
	if s.geofence != nil && s.geofence.IsInside(loc.Latitude, loc.Longitude) {
		event := models.GeofenceEvent{
			VehicleID: loc.VehicleID,
			Event:     "geofence_entry",
			Timestamp: loc.Timestamp,
		}
		event.Location.Latitude = loc.Latitude
		event.Location.Longitude = loc.Longitude

		if err := s.rabbitCli.PublishGeofenceEvent(ctx, event); err != nil {
			log.Printf("failed to publish geofence event: %v", err)
		} else {
			log.Printf("Published geofence_entry for %s", loc.VehicleID)
		}
	}

	return nil
}

func (s *LocationService) GetLatest(ctx context.Context, vehicleID string) (*models.VehicleLocation, error) {
	return s.repo.GetLatest(ctx, vehicleID)
}

func (s *LocationService) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]models.VehicleLocation, error) {
	return s.repo.GetHistory(ctx, vehicleID, start, end)
}
