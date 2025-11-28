package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppPort string

	PostgresURL string

	MQTTBrokerURL string
	MQTTClientID  string

	RabbitURL        string
	RabbitExchange   string
	RabbitQueue      string
	RabbitRoutingKey string

	// Geofence
	GeofenceLat    float64
	GeofenceLon    float64
	GeofenceRadius float64 // in meters
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
		log.Printf("WARN: invalid float for %s: %v", key, err)
	}
	return def
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),

		PostgresURL: getEnv("POSTGRES_URL", "postgres://mastama:post456@db:5432/fleetdb?sslmode=disable"),

		MQTTBrokerURL: getEnv("MQTT_BROKER_URL", "tcp://mqtt:1883"),
		MQTTClientID:  getEnv("MQTT_CLIENT_ID", "fleet-backend"),

		RabbitURL:        getEnv("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/"),
		RabbitExchange:   getEnv("RABBIT_EXCHANGE", "fleet.events"),
		RabbitQueue:      getEnv("RABBIT_QUEUE", "geofence_alerts"),
		RabbitRoutingKey: getEnv("RABBIT_ROUTING_KEY", "geofence.entry"),

		GeofenceLat:    getEnvFloat("GEOFENCE_LAT", -6.2088),
		GeofenceLon:    getEnvFloat("GEOFENCE_LON", 106.8456),
		GeofenceRadius: getEnvFloat("GEOFENCE_RADIUS", 50), // 50 meter
	}
}
