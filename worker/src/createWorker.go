package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Worker struct {
	db *sql.DB
}

func connectPostgres() (*sql.DB, error) {
	// read environment variables
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	for i := 1; i <= 10; i++ {
		db, err := sql.Open("pgx", connStr)

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
	db, err := connectPostgres()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Postgres: %w", err)
	}

	// Test connection
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx2); err != nil {
		return nil, fmt.Errorf("cannot ping Postgres: %w", err)
	}

	return &Worker{
		db: db,
	}, nil
}
