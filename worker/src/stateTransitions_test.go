package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/me/job-processor/worker/src/testutil"
	"github.com/stretchr/testify/require"
)

func TestQueuePendingJobs(t *testing.T) {
	db, ctx := refreshDB(t)

	// Seed data
	_, err := db.Exec(`
		INSERT INTO jobs (description, status)
		VALUES
		  ('a', 'PENDING'),
		  ('b', 'PENDING'),
		  ('c', 'RUNNING')
	`)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	ids, err := queuePendingJobs(ctx, tx)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	require.NoError(t, tx.Commit())

	// Assert DB state
	rows, err := db.Query(`
		SELECT status FROM jobs ORDER BY id
	`)
	require.NoError(t, err)

	var statuses []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		statuses = append(statuses, s)
	}

	require.Equal(t,
		[]string{"QUEUED", "QUEUED", "RUNNING"},
		statuses,
	)
}

func TestRequeueStuckRunningJobs(t *testing.T) {
	db, ctx := refreshDB(t)

	// Seed data, interval must be greater that worker_running_timeout
	_, err := db.Exec(`
		INSERT INTO jobs (description, status, started_at)
		VALUES
		('a', 'RUNNING', NOW() - INTERVAL '1 DAY'),
		('b', 'RUNNING', NOW() - INTERVAL '5 MINUTES'),
		('c', 'RUNNING', NOW())
	`)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	ids, err := requeueStuckRunningJobs(ctx, tx)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	require.NoError(t, tx.Commit())

	// Assert DB state
	rows, err := db.Query(`
		SELECT status FROM jobs ORDER BY id
	`)
	require.NoError(t, err)

	var statuses []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		statuses = append(statuses, s)
	}

	require.Equal(t,
		[]string{"QUEUED", "RUNNING", "RUNNING"},
		statuses,
	)
}

func TestClaimJob(t *testing.T) {
	db, ctx := refreshDB(t)

	rows, err := db.QueryContext(ctx, `
		INSERT INTO jobs (description, status)
		VALUES
			('a', 'QUEUED'),
			('b', 'QUEUED'),
			('c', 'QUEUED')
		RETURNING id
	`)
	require.NoError(t, err)

	rows.Next()
	var id int
	rows.Scan(&id)
	require.Equal(t, 1, id)

	w, err := NewWorker()
	require.NoError(t, err)

	err = w.claimJob(ctx, id)
	require.NoError(t, err)

	// Assert DB state
	rows, err = db.Query(`
		SELECT status FROM jobs ORDER BY id
	`)
	require.NoError(t, err)

	var statuses []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		statuses = append(statuses, s)
	}

	require.Equal(t,
		[]string{"RUNNING", "QUEUED", "QUEUED"},
		statuses,
	)
}

func TestHandleJobFailure(t *testing.T) {
	db, ctx := refreshDB(t)
	maxAttempts, err := strconv.Atoi(os.Getenv("worker_max_job_retries"))
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, `
		INSERT INTO jobs (description, status, attempts)
		VALUES
			('a', 'RUNNING', $1),
			('b', 'RUNNING', $2),
			('c', 'RUNNING', $3)
		RETURNING id
	`, maxAttempts-2, maxAttempts-1, maxAttempts)
	require.NoError(t, err)

	var jobIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			jobIDs = append(jobIDs, id)
		}
	}

	w, err := NewWorker()
	require.NoError(t, err)
	err = errors.New("a basic error message")
	for _, id := range jobIDs {
		w.handleJobFailure(ctx, id, err)
	}

	// Assert DB state
	rows, err = db.Query(`
		SELECT status FROM jobs ORDER BY id
	`)
	require.NoError(t, err)

	var statuses []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		statuses = append(statuses, s)
	}

	require.Equal(t,
		[]string{"PENDING", "FAILED", "FAILED"},
		statuses,
	)
}

func refreshDB(t *testing.T) (*sql.DB, context.Context) {
	db := testutil.OpenTestDB(t)
	testutil.TruncateJobs(t, db)
	ctx := context.Background()
	return db, ctx
}
