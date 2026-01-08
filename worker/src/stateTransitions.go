package main

import (
	"context"
	"database/sql"
	"log"
)

const maxAttempts = 3

// Queue jobs in QUEUED state in order to recover any lost jobs due to crashes, idempotent worker prevents double processing
func queueQueuedJobs(ctx context.Context, tx *sql.Tx) ([]int, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE jobs
		SET enqueued_at = NOW()
		WHERE id IN (
			SELECT id
			FROM jobs
			WHERE status = $1
			AND (enqueued_at IS NULL OR enqueued_at < NOW() - INTERVAL '5 minutes')
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id;
	`, StatusQueued)
	if err != nil {
		return nil, err
	}

	var jobIDs = rowsToIDs(rows)

	return jobIDs, nil
}

// UPDATE and get IDs atomically
func queuePendingJobs(ctx context.Context, tx *sql.Tx) ([]int, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE jobs
		SET status = $1,
		enqueued_at = NOW()
		WHERE id IN (
			SELECT id
			FROM jobs
			WHERE status = $2
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 10
		)
		RETURNING id
	`, StatusQueued, StatusPending)
	if err != nil {
		return nil, err
	}

	var jobIDs = rowsToIDs(rows)

	return jobIDs, nil
}

// Requeue stuck RUNNING jobs that have exceeded the time limit
func requeueStuckRunningJobs(ctx context.Context, tx *sql.Tx) ([]int, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE jobs
		SET status = $1,
		started_at = NULL,
		enqueued_at = NOW()
		WHERE id IN (
			SELECT id
			FROM jobs
			WHERE status=$2
			AND started_at < NOW() - INTERVAL '1 hour'
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 10
		)
		RETURNING id
	`, StatusQueued, StatusRunning)
	if err != nil {
		return nil, err
	}

	var jobIDs = rowsToIDs(rows)

	return jobIDs, nil
}

func (w *Worker) claimJob(ctx context.Context, jobID int) error {
	res, err := w.db.ExecContext(
		ctx,
		"UPDATE jobs SET status=$1, started_at=NOW() WHERE id=$2 AND status=$3",
		StatusRunning,
		jobID,
		StatusQueued,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (w *Worker) handleJobFailure(ctx context.Context, jobID int, err error) {
	var attempts int

	//set to requeued or failed based on attempts < max_attempts
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
    `, jobID, err.Error(), maxAttempts, StatusFailed, StatusQueued, StatusRunning).Scan(&attempts)

	if err2 != nil {
		log.Printf("Failed to update retry state for job %d: %v", jobID, err2)
	}
}

func rowsToIDs(rows *sql.Rows) []int {
	defer rows.Close()

	var jobIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			jobIDs = append(jobIDs, id)
		}
	}

	return jobIDs
}
