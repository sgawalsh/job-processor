package main

import (
	"context"
	"database/sql"
	"log"
	"os"
)

var maxAttempts = os.Getenv("worker_max_job_retries")
var runningTimeout = os.Getenv("worker_running_timeout")

// Requeue stuck RUNNING jobs that have exceeded the time limit
func requeueStuckRunningJobs(ctx context.Context, db *sql.DB) ([]int, error) {
	rows, err := db.QueryContext(ctx, `
		UPDATE jobs
		SET status = $1,
		started_at = NULL
		WHERE id IN (
			SELECT id
			FROM jobs
			WHERE status=$2
			AND started_at < NOW() - ($3)::interval
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 10
		)
		RETURNING id
	`, StatusPending, StatusRunning, runningTimeout)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			jobIDs = append(jobIDs, id)
		}
	}

	return jobIDs, nil
}

func (w *Worker) claimNextJob(ctx context.Context) (int, error) {
	var jobID int
	err := w.db.QueryRowContext(ctx, `
		UPDATE jobs
		SET status = $1,
		    started_at = NOW()
		WHERE id IN (
			SELECT id
			FROM jobs
			WHERE status = $2
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id
	`, StatusRunning, StatusPending).Scan(&jobID)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, sql.ErrNoRows
		}
		return 0, err
	}

	return jobID, nil
}

func (w *Worker) handleJobFailure(ctx context.Context, jobID int, err error) {
	var attempts int

	// increment attempts counter and set to pending or failed based on attempts < max_attempts
	err2 := w.db.QueryRowContext(ctx, `
        UPDATE jobs
        SET attempts = attempts + 1,
            last_error = $2,
            status = CASE
                WHEN attempts + 1 >= $3 THEN $4::job_status
                ELSE $5::job_status
            END
        WHERE id = $1
		AND status = $6
        RETURNING attempts
    `, jobID, err.Error(), maxAttempts, StatusFailed, StatusPending, StatusRunning).Scan(&attempts)

	if err2 != nil {
		log.Printf("Failed to update retry state for job %d: %v", jobID, err2)
	}
}
