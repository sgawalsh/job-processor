package testutil

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "postgres://test:test@localhost:5432/test?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TruncateJobs(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE jobs RESTART IDENTITY`)
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}
