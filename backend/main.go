package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "oskar/cwm25/docs"
	"oskar/cwm25/internal"
)

//	@title			CWM Kanban
//	@version		0.2
//	@description	This is a small kanban server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath	/
func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL env var not set")
	}

	// Set up the DB connection. If it is not up yet, wait up to 15s and retry.
	var db *internal.DB
	var err error
	var wait time.Duration
	maxAttempts := 5
	for attempt := range maxAttempts {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err = internal.NewDB(ctx, connString)
		if err == nil {
			defer cancel()
			break
		}
		cancel()

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
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Println("SECRET env var not set; falling back to default")
		secret = "super-secure-hardcoded-secret-for-testing-purposes-only"
	}
	addr := fmt.Sprintf(":%s", port)
	handler := internal.PrepareServer(addr, secret, db)

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
