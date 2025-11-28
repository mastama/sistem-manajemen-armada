// cmd/mqtt-listener/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/geofence"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type VehicleLocation struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type GeofenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"`
	Location  struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Timestamp int64 `json:"timestamp"`
}

// --- helper retry Postgres ---
func waitForPostgres(ctx context.Context, dsn string, maxAttempts int, baseDelay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			log.Printf("postgres create pool try %d failed: %v", attempt, err)
			lastErr = err
		} else {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err = pool.Ping(pingCtx)
			cancel()

			if err == nil {
				log.Println("MQTT listener connected to PostgreSQL")
				return pool, nil
			}

			log.Printf("postgres ping try %d failed: %v", attempt, err)
			lastErr = err
			pool.Close()
		}

		time.Sleep(time.Duration(attempt) * baseDelay)
	}

	return nil, fmt.Errorf("failed to connect to postgres after %d attempts: %w", maxAttempts, lastErr)
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// --- Inisialisasi geofence sekali, dipakai di handler ---
	gf := geofence.NewGeofence(cfg.GeofenceLat, cfg.GeofenceLon, cfg.GeofenceRadius)

	// --- Postgres dengan retry ---
	dbpool, err := waitForPostgres(ctx, cfg.PostgresURL, 10, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	// --- RabbitMQ ---
	var conn *amqp.Connection
	for i := 1; i <= 10; i++ {
		conn, err = amqp.Dial(cfg.RabbitURL)
		if err == nil {
			break
		}
		log.Printf("rabbitmq connect try %d failed: %v", i, err)
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel error: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		cfg.RabbitExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange declare: %v", err)
	}

	q, err := ch.QueueDeclare(
		cfg.RabbitQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("queue declare: %v", err)
	}

	if err := ch.QueueBind(
		q.Name,
		cfg.RabbitRoutingKey,
		cfg.RabbitExchange,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue bind: %v", err)
	}

	// --- MQTT ---
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL).
		SetClientID("fleet-mqtt-listener")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt connect error: %v", token.Error())
	}
	log.Println("MQTT listener connected to MQTT broker")

	topic := "/fleet/vehicle/+/location"

	handler := func(c mqtt.Client, m mqtt.Message) {
		log.Printf("Received on %s: %s", m.Topic(), string(m.Payload()))

		var loc VehicleLocation
		if err := json.Unmarshal(m.Payload(), &loc); err != nil {
			log.Printf("invalid JSON: %v", err)
			return
		}

		// Jika vehicle_id kosong, ambil dari topic: /fleet/vehicle/{id}/location
		if loc.VehicleID == "" {
			parts := strings.Split(m.Topic(), "/")
			if len(parts) >= 4 {
				loc.VehicleID = parts[3]
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := dbpool.Exec(ctx,
			`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
             VALUES ($1, $2, $3, $4)`,
			loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp,
		)
		if err != nil {
			log.Printf("insert failed: %v", err)
			return
		}

		// Pakai geofence.IsInside dari package internal/geofence
		if gf.IsInside(loc.Latitude, loc.Longitude) {
			ev := GeofenceEvent{
				VehicleID: loc.VehicleID,
				Event:     "geofence_entry",
				Timestamp: loc.Timestamp,
			}
			ev.Location.Latitude = loc.Latitude
			ev.Location.Longitude = loc.Longitude

			body, _ := json.Marshal(ev)
			if err := ch.PublishWithContext(ctx,
				cfg.RabbitExchange,
				cfg.RabbitRoutingKey,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			); err != nil {
				log.Printf("publish geofence event failed: %v", err)
			} else {
				log.Printf("Published geofence event for %s", loc.VehicleID)
			}
		}
	}

	if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt subscribe error: %v", token.Error())
	}
	log.Printf("Subscribed to MQTT topic %s", topic)

	select {}
}
