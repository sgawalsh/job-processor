package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	//Start poller
	wg.Go(func() {
		w.pollPendingJobs(ctx)
	})

	//Start Redis consumer
	wg.Go(func() {
		w.executeQueuedJobs(ctx)
	})

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
