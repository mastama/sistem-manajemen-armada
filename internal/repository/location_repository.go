package repository

import (
	"context"

	"sistem-manajemen-armada/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationRepository interface {
	Insert(ctx context.Context, loc models.VehicleLocation) error
	GetLatest(ctx context.Context, vehicleID string) (*models.VehicleLocation, error)
	GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]models.VehicleLocation, error)
}

type locationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) Insert(ctx context.Context, loc models.VehicleLocation) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		 VALUES ($1, $2, $3, $4)`,
		loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp,
	)
	return err
}

func (r *locationRepository) GetLatest(ctx context.Context, vehicleID string) (*models.VehicleLocation, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, vehicle_id, latitude, longitude, timestamp
		 FROM vehicle_locations
		 WHERE vehicle_id = $1
		 ORDER BY timestamp DESC
		 LIMIT 1`,
		vehicleID,
	)

	var loc models.VehicleLocation
	if err := row.Scan(&loc.ID, &loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *locationRepository) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]models.VehicleLocation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, vehicle_id, latitude, longitude, timestamp
		 FROM vehicle_locations
		 WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		 ORDER BY timestamp ASC`,
		vehicleID, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.VehicleLocation
	for rows.Next() {
		var loc models.VehicleLocation
		if err := rows.Scan(&loc.ID, &loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, err
		}
		result = append(result, loc)
	}
	return result, nil
}
