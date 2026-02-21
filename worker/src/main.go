package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	jobsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_jobs_processed_total",
			Help: "Total jobs processed",
		},
	)
)

var (
	jobsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_jobs_failed_total",
			Help: "Total jobs failed",
		},
	)
)

func main() {
	log.Println("Worker starting...")

	// Create a new worker
	w, err := NewWorker()
	if err != nil {
		log.Fatalf("Failed to initialize worker: %v", err)
	}

	// Root cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Go(func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	})

	switch os.Getenv("ROLE") {
	case "poller": // Start poller
		wg.Go(func() {
			w.pollPendingJobs(ctx)
		})
	case "worker": // Start DB jobs consumer
		prometheus.MustRegister(jobsProcessed)
		prometheus.MustRegister(jobsFailed)
		wg.Go(func() {
			w.executeDBJobs(ctx)
		})
	default:
		log.Fatal("ROLE must be poller or worker")
	}

	// Handle shutdown signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("Shutdown signal received")

	cancel()

	// Forced shutdown after timeout
	const maxShutdown = 10 * time.Second
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers shut down cleanly")
	case <-time.After(maxShutdown):
		log.Println("Shutdown timed out, forcing exit")
	}

	log.Println("Worker stopped")
}
