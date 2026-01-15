package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient *redis.Client
	db          *sql.DB
}

const (
	StatusPending   = "PENDING"
	StatusQueued    = "QUEUED"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
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
	}, nil
}

// Run starts consuming jobs from Redis
func (w *Worker) executeQueuedJobs(ctx context.Context) {
	log.Println("Worker is running...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Received stop signal")
			return
		default:
			// Wait for job ID from Redis queue
			result, err := w.redisClient.BLPop(ctx, 5*time.Second, "jobs:queue").Result()
			if err != nil {
				if err == redis.Nil {
					// Queue was empty, nothing to process, no log needed
					continue
				}
				log.Printf("Error fetching job: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			jobIDStr := result[1]
			jobID, err := strconv.Atoi(jobIDStr)
			if err != nil {
				log.Printf("Invalid jobID in queue: %q", jobIDStr)
				continue // poison message, skip
			}

			// Fetch job from Postgres
			var description, status string
			err = w.db.QueryRowContext(ctx, "SELECT description, status FROM jobs WHERE id=$1", jobID).Scan(&description, &status)
			if err != nil {
				log.Printf("Error fetching job from DB: %v", err)
				continue
			}

			log.Printf("Found job %d: %s (current status: %s)", jobID, description, status)

			err = w.claimJob(ctx, jobID)
			if err != nil {
				log.Printf("Error claiming job %d: %v", jobID, err)
				continue
			}

			log.Printf("Processing job: %d", jobID)

			err = doWork()
			if err != nil {
				w.handleJobFailure(ctx, jobID, err)
				continue
			}

			// Update job status to "completed"
			_, err = w.db.ExecContext(ctx, "UPDATE jobs SET status=$1 WHERE id=$2 and status=$3", StatusSucceeded, jobID, StatusRunning)
			if err != nil {
				log.Printf("Error updating job %d to completed status: %v", jobID, err)
				continue
			}
			jobsProcessed.Inc()
			log.Printf("Job %d marked as completed", jobID)
		}
	}
}

func (w *Worker) pollPendingJobs(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Poller stopping...")
			return
		case <-ticker.C:
			tx, err := w.db.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("Poller: failed to begin tx: %v", err)
				continue
			}

			// Queue jobs in QUEUED state in order to recover any lost jobs due to crashes, idempotent worker prevents double processing
			jobIDs, err := queueQueuedJobs(ctx, tx)
			if err != nil {
				log.Printf("Poller: failed to enqueue queued jobs: %v", err)
				tx.Rollback()
				continue
			}

			// Set PENDING jobs to QUEUED in DB and get their IDs
			pendingJobIDs, err := queuePendingJobs(ctx, tx)
			if err != nil {
				log.Printf("Poller: failed to enqueue pending jobs: %v", err)
				tx.Rollback()
				continue
			}

			// Requeue stuck RUNNING jobs
			stuckJobIDs, err := requeueStuckRunningJobs(ctx, tx)
			if err != nil {
				log.Printf("Poller: failed to requeue stuck running jobs: %v", err)
				tx.Rollback()
				continue
			}

			if err := tx.Commit(); err != nil {
				log.Printf("Poller: commit failed: %v", err)
				continue
			}

			jobIDs = slices.Concat(jobIDs, pendingJobIDs, stuckJobIDs)

			// Enqueue committed jobs to Redis
			for _, id := range jobIDs {
				if err := w.redisClient.RPush(ctx, "jobs:queue", id).Err(); err != nil {
					log.Printf("Poller: failed to enqueue job %d: %v", id, err)
				}
			}
			if len(jobIDs) > 0 {
				log.Printf("Poller: enqueued %d jobs", len(jobIDs))
			}

		}
	}
}

func doWork() error {
	// Simulate work
	time.Sleep(2 * time.Second)
	return nil
}
