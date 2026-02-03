package main

import (
	"context"
	"log"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StatusPending   = "PENDING"
	StatusQueued    = "QUEUED"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
)

var pollingInterval = os.Getenv("worker_poll_interval")
var workerTimeoutInterval = os.Getenv("worker_execution_timeout_interval")

// Run starts consuming jobs from Redis
func (w *Worker) executeQueuedJobs(ctx context.Context) {
	log.Println("Worker is running...")

	interval, err := strconv.Atoi(workerTimeoutInterval)
	if err != nil {
		log.Fatalf("Failed to parse worker_execution_timeout_interval: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Received stop signal")
			return
		default:
			
			// Wait for job ID from Redis queue
			result, err := w.redisClient.BLPop(ctx, time.Duration(interval)*time.Second, "jobs:queue").Result()
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
				jobsFailed.Inc()
				log.Printf("Job %d failed: %v", jobID, err)
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
	interval, err := strconv.Atoi(pollingInterval)
	if err != nil {
		log.Printf("Failed to parse worker_poll_interval: %v", err)
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
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
