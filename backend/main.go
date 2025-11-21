package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"oskar/cwm25/internal"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL env var not set")
	}

	// Set up the DB connection. If it is not up yet, wait up to 15s and retry.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var db *internal.DB
	var err error
	var wait time.Duration
	maxAttempts := 5
	for attempt := range maxAttempts {
		db, err = internal.NewDB(ctx, connString)
		if err == nil {
			break
		}

		wait = time.Second * time.Duration(attempt+1)
		log.Printf("failed to connect to db (attempt %d/%d): %v; retrying in %s", attempt, maxAttempts, err, wait)
		time.Sleep(wait)
	}
	if err != nil {
		log.Fatalf("failed to connect to db after %d attempts: %v", maxAttempts, err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	handler := internal.PrepareServer(addr, db)

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
