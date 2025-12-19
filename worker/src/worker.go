package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient *redis.Client
	db          *sql.DB
	quit        chan bool
}

const (
	StatusPending    = 0
	StatusProcessing = 1
	StatusCompleted  = 2
	StatusFailed     = 3
)

func connectPostgres() (*sql.DB, error) {
	for i := 1; i <= 10; i++ {
		db, err := sql.Open("pgx", "postgres://app:app@db:5432/jobs?sslmode=disable")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Println("Connected to Postgres")
			return db, nil
		}

		log.Printf("Waiting for Postgres (%d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to Postgres after retries")
}

// NewWorker initializes the Redis client
func NewWorker() (*Worker, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Docker Compose service name
		Password: "",           // No password for local dev
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cannot connect to Redis: %w", err)
	}

	db, err := connectPostgres()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Postgres: %w", err)
	}

	// Test connection
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := db.PingContext(ctx2); err != nil {
		return nil, fmt.Errorf("cannot ping Postgres: %w", err)
	}

	return &Worker{
		redisClient: rdb,
		db:          db,
		quit:        make(chan bool),
	}, nil
}

// Run starts consuming jobs from Redis
func (w *Worker) Run() {
	ctx := context.Background()
	log.Println("Worker is running...")

	for {
		select {
		case <-w.quit:
			log.Println("Received stop signal")
			return
		default:
			// Wait for job ID from Redis queue
			result, err := w.redisClient.BLPop(ctx, 0*time.Second, "jobs:queue").Result()
			if err != nil {
				log.Printf("Error fetching job: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			jobID := result[1]
			log.Printf("Processing job: %s", jobID)

			// Fetch job from Postgres
			var description, status string
			err = w.db.QueryRowContext(ctx, "SELECT description, status FROM jobs WHERE id=$1", jobID).Scan(&description, &status)
			if err != nil {
				log.Printf("Error fetching job from DB: %v", err)
				continue
			}

			log.Printf("Job %s: %s (current status: %s)", jobID, description, status)

			_, err = w.db.ExecContext(
				ctx,
				"UPDATE jobs SET status=$1 WHERE id=$2 AND status=$3",
				StatusProcessing,
				jobID,
				StatusPending,
			)

			// Simulate work
			time.Sleep(2 * time.Second)

			// Update job status to "completed"
			_, err = w.db.ExecContext(ctx, "UPDATE jobs SET status=$1 WHERE id=$2", StatusCompleted, jobID)
			if err != nil {
				log.Printf("Error updating job status: %v", err)
				continue
			}

			log.Printf("Job %s marked as completed", jobID)
		}
	}
}

// Stop signals the worker to stop
func (w *Worker) Stop() {
	close(w.quit)
}
