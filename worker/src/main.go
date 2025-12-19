package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    log.Println("Worker starting...")

    // Create a new worker
    w, err := NewWorker()
    if err != nil {
        log.Fatalf("Failed to initialize worker: %v", err)
    }

    // Run worker in a separate goroutine
    done := make(chan bool)
    go func() {
        w.Run()
        done <- true
    }()

    // Wait for SIGINT or SIGTERM to gracefully shut down
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    <-sigs
    log.Println("Shutting down worker...")
    w.Stop()
    <-done
    log.Println("Worker stopped.")
}