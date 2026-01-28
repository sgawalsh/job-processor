package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient *redis.Client
	db          *sql.DB
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
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
