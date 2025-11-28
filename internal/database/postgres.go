package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(url string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("failed to parse postgres url: %v", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2

	var pool *pgxpool.Pool

	// Coba connect + ping sampai 10x dengan backoff
	for i := 1; i <= 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			// ping untuk memastikan DB ready
			if pingErr := pool.Ping(ctx); pingErr == nil {
				cancel()
				log.Println("Connected to PostgreSQL")
				return pool
			} else {
				log.Printf("postgres ping try %d failed: %v", i, pingErr)
				pool.Close()
			}
		} else {
			log.Printf("postgres connect try %d failed: %v", i, err)
		}

		cancel()
		// backoff incremental
		sleep := time.Duration(i) * time.Second
		log.Printf("retrying postgres in %s...", sleep)
		time.Sleep(sleep)
	}

	log.Fatalf("failed to connect to postgres after retries: %v", err)
	return nil
}
