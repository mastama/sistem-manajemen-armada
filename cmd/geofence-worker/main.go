package main

import (
	"encoding/json"
	"log"
	"time"

	"sistem-manajemen-armada/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type GeofenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"`
	Location  struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Timestamp int64 `json:"timestamp"`
}

func main() {
	cfg := config.Load()

	var conn *amqp.Connection
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = amqp.Dial(cfg.RabbitURL)
		if err == nil {
			break
		}
		log.Printf("rabbitmq connect try %d failed: %v", i, err)
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect rabbitmq after retries: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel error: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		cfg.RabbitExchange, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,
	); err != nil {
		log.Fatalf("exchange declare error: %v", err)
	}

	q, err := ch.QueueDeclare(
		cfg.RabbitQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("queue declare error: %v", err)
	}

	if err := ch.QueueBind(
		q.Name,
		cfg.RabbitRoutingKey,
		cfg.RabbitExchange,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue bind error: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,  // auto-ack
		false, // exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume error: %v", err)
	}

	log.Printf("Geofence worker listening on queue: %s", q.Name)

	for msg := range msgs {
		var event GeofenceEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("invalid event: %v", err)
			continue
		}

		log.Printf(
			"[GEOFENCE ALERT] vehicle=%s event=%s lat=%.6f lon=%.6f ts=%d",
			event.VehicleID,
			event.Event,
			event.Location.Latitude,
			event.Location.Longitude,
			event.Timestamp,
		)
	}
}
