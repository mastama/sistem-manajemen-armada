package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
}

func NewClient(cfg *config.Config) *Client {
	var (
		conn *amqp.Connection
		ch   *amqp.Channel
		err  error
	)

	// Retry connect ke RabbitMQ (maks 10 kali dengan backoff)
	for i := 1; i <= 10; i++ {
		conn, err = amqp.Dial(cfg.RabbitURL)
		if err == nil {
			ch, err = conn.Channel()
			if err == nil {
				// Berhasil buka channel, break dari loop
				break
			}
			// gagal buka channel → tutup koneksi dulu
			_ = conn.Close()
		}

		log.Printf("rabbitmq connect try %d failed: %v", i, err)
		sleep := time.Duration(i) * time.Second
		log.Printf("retrying rabbitmq in %s...", sleep)
		time.Sleep(sleep)
	}

	if err != nil {
		log.Fatalf("failed to connect to rabbitmq after retries: %v", err)
	}

	// Pastikan exchange + queue + binding ada
	if err := ch.ExchangeDeclare(
		cfg.RabbitExchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	if _, err := ch.QueueDeclare(
		cfg.RabbitQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false,
		nil,
	); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	if err := ch.QueueBind(
		cfg.RabbitQueue,
		cfg.RabbitRoutingKey,
		cfg.RabbitExchange,
		false,
		nil,
	); err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	log.Println("Connected to RabbitMQ")

	return &Client{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}
}

func (c *Client) PublishGeofenceEvent(ctx context.Context, event models.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return c.channel.PublishWithContext(
		ctx,
		c.cfg.RabbitExchange,
		c.cfg.RabbitRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (c *Client) Channel() *amqp.Channel {
	return c.channel
}
