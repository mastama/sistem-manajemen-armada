package main

import (
	"log"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/database"
	"sistem-manajemen-armada/internal/geofence"
	httpHandler "sistem-manajemen-armada/internal/http"
	"sistem-manajemen-armada/internal/rabbitmq"
	"sistem-manajemen-armada/internal/repository"
	"sistem-manajemen-armada/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := database.NewPostgresPool(cfg.PostgresURL)
	defer db.Close()

	gf := geofence.NewGeofence(cfg.GeofenceLat, cfg.GeofenceLon, cfg.GeofenceRadius)
	rabbit := rabbitmq.NewClient(cfg)

	repo := repository.NewLocationRepository(db)
	svc := service.NewLocationService(repo, gf, rabbit)

	r := gin.Default()
	h := httpHandler.NewHandler(svc)
	h.RegisterRoutes(r)

	log.Printf("API server listening on :%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("api server failed: %v", err)
	}
}
