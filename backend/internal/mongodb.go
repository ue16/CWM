package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var _ Store = (*MongoDB)(nil)

// MongoDB wraps the mongo-go-driver and provides CRUD helpers for the Kanban models.
type MongoDB struct {
	client *mongo.Client
	cancel context.CancelFunc
	db     *mongo.Database
}

// NewMongoDB creates a new DB given a MongoDB connection string. It retries for up
// to 5 times (max ~1 Minute) or fails.
// Example connString: "mongodb://user:pass@localhost:27017/dbname"
func NewMongoDB(connString string) (*MongoDB, error) {
	connect := func(ctx context.Context) (*MongoDB, error) {
		client, err := mongo.Connect(options.Client().ApplyURI(connString))
		if err != nil {
			return nil, fmt.Errorf("mongo.Connect: %w", err)
		}

		// Verify connection
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			client.Disconnect(ctx)
			return nil, fmt.Errorf("ping db: %w", err)
		}

		return &MongoDB{client: client}, nil
	}

	var db *MongoDB
	var err error
	var wait time.Duration
	maxAttempts := 5
	for attempt := range maxAttempts {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err = connect(ctx)
		if err == nil {
			db.cancel = cancel
			splitConnString := strings.Split(connString, "/")
			if len(splitConnString) < 4 {
				cancel()
				return nil, fmt.Errorf("mongodb connection string does not contain database name")
			}
			db.db = db.client.Database(splitConnString[3])
			break
		}
		cancel()

		wait = time.Second * time.Duration(attempt+1)
		log.Printf("failed to connect to db (attempt %d/%d): %v; retrying in %s", attempt, maxAttempts, err, wait)
		time.Sleep(wait)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db after %d attempts: %v", maxAttempts, err)
	}
	return db, nil
}

func (db *MongoDB) Close() {
	db.cancel()
	db.client.Disconnect(context.Background())
}
