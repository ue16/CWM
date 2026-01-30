package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
			splitConnString := strings.Split(strings.Split(connString, "?")[0], "/")
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

// getNextID gets the next ID for a collection
func (db *MongoDB) getNextID(ctx context.Context, collectionName string) (int, error) {
	counters := db.db.Collection("counters")
	result := counters.FindOneAndUpdate(
		ctx,
		bson.M{"_id": collectionName},
		bson.M{"$inc": bson.M{"sequence": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var counter bson.M
	if err := result.Decode(&counter); err != nil {
		// If document doesn't exist, create it
		if err == mongo.ErrNoDocuments {
			counters.InsertOne(ctx, bson.M{
				"_id":      collectionName,
				"sequence": 1,
			})
			return 1, nil
		}
		return 0, err
	}

	// Handle both int32 and int64 types from MongoDB
	seq := counter["sequence"]
	switch v := seq.(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("getNextID: sequence value has unexpected type %T", seq)
	}
}

//
// Users
//

// CreateUser inserts a new user and returns it with the assigned ID.
func (db *MongoDB) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
	id, err := db.getNextID(ctx, "users")
	if err != nil {
		return User{}, fmt.Errorf("CreateUser: getNextID: %w", err)
	}

	user := User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	users := db.db.Collection("users")
	_, err = users.InsertOne(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("CreateUser: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username.
func (db *MongoDB) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	users := db.db.Collection("users")
	err := users.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, fmt.Errorf("GetUserByUsername: not found")
		}
		return User{}, fmt.Errorf("GetUserByUsername: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by id.
func (db *MongoDB) GetUserByID(ctx context.Context, id int) (User, error) {
	var u User
	users := db.db.Collection("users")
	err := users.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, fmt.Errorf("GetUserByID: not found")
		}
		return User{}, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

// UserExists checks if a user with the given username exists.
func (db *MongoDB) UserExists(ctx context.Context, username string) (bool, error) {
	users := db.db.Collection("users")
	result := users.FindOne(ctx, bson.M{"username": username})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("UserExists: %w", result.Err())
	}
	return true, nil
}

//
// Boards
//

// CreateBoard inserts a new board and returns it with the assigned ID.
func (db *MongoDB) CreateBoard(ctx context.Context, name string, userID int) (Board, error) {
	id, err := db.getNextID(ctx, "boards")
	if err != nil {
		return Board{}, fmt.Errorf("CreateBoard: getNextID: %w", err)
	}

	board := Board{
		ID:     id,
		Name:   name,
		UserID: userID,
	}

	boards := db.db.Collection("boards")
	_, err = boards.InsertOne(ctx, board)
	if err != nil {
		return Board{}, fmt.Errorf("CreateBoard: %w", err)
	}

	return board, nil
}

// GetBoardByID retrieves a board by id.
func (db *MongoDB) GetBoardByID(ctx context.Context, id int) (Board, error) {
	var b Board
	boards := db.db.Collection("boards")
	err := boards.FindOne(ctx, bson.M{"id": id}).Decode(&b)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Board{}, fmt.Errorf("GetBoardByID: not found")
		}
		return Board{}, fmt.Errorf("GetBoardByID: %w", err)
	}
	return b, nil
}

// ListBoards returns all boards for a user.
func (db *MongoDB) ListBoards(ctx context.Context, userID int) ([]Board, error) {
	boards := db.db.Collection("boards")
	cursor, err := boards.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("ListBoards: %w", err)
	}
	defer cursor.Close(ctx)

	var result []Board
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("ListBoards: %w", err)
	}

	return result, nil
}

// UpdateBoard updates a board's name.
func (db *MongoDB) UpdateBoard(ctx context.Context, b Board) error {
	boards := db.db.Collection("boards")
	result, err := boards.UpdateOne(ctx, bson.M{"id": b.ID}, bson.M{"$set": b})
	if err != nil {
		return fmt.Errorf("UpdateBoard: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("UpdateBoard: not found")
	}
	return nil
}

// DeleteBoard deletes a board by id. Cascade deletes columns and cards.
func (db *MongoDB) DeleteBoard(ctx context.Context, id int) error {
	boards := db.db.Collection("boards")
	result, err := boards.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("DeleteBoard: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("DeleteBoard: not found")
	}

	// Cascade delete: delete all columns and their cards
	columns := db.db.Collection("columns")
	colResult, err := columns.DeleteMany(ctx, bson.M{"board_id": id})
	if err != nil {
		return fmt.Errorf("DeleteBoard: cascade delete columns: %w", err)
	}

	// Delete all cards and attachments in those columns
	if colResult.DeletedCount > 0 {
		cards := db.db.Collection("cards")
		_, err = cards.DeleteMany(ctx, bson.M{"board_id": id})
		if err != nil {
			return fmt.Errorf("DeleteBoard: cascade delete cards: %w", err)
		}

		attachments := db.db.Collection("attachments")
		_, err = attachments.DeleteMany(ctx, bson.M{"board_id": id})
		if err != nil {
			return fmt.Errorf("DeleteBoard: cascade delete attachments: %w", err)
		}
	}

	return nil
}

// CompareBoardOwner compares the owner of a board with a given user ID. Returns true if the owner matches the user ID.
func (db *MongoDB) CompareBoardOwner(ctx context.Context, id, user_id int) (bool, error) {
	var b Board
	boards := db.db.Collection("boards")
	err := boards.FindOne(ctx, bson.M{"id": id}).Decode(&b)
	if err != nil {
		return false, fmt.Errorf("CompareBoardOwner: %w", err)
	}
	return b.UserID == user_id, nil
}

//
// Columns
//

// CreateColumn inserts a new column and returns it with the assigned ID.
func (db *MongoDB) CreateColumn(ctx context.Context, c Column) (Column, error) {
	id, err := db.getNextID(ctx, "columns")
	if err != nil {
		return Column{}, fmt.Errorf("CreateColumn: getNextID: %w", err)
	}

	c.ID = id
	columns := db.db.Collection("columns")
	_, err = columns.InsertOne(ctx, c)
	if err != nil {
		return Column{}, fmt.Errorf("CreateColumn: %w", err)
	}

	return c, nil
}

// GetColumnByID retrieves a column by id.
func (db *MongoDB) GetColumnByID(ctx context.Context, id int) (Column, error) {
	var c Column
	columns := db.db.Collection("columns")
	err := columns.FindOne(ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Column{}, fmt.Errorf("GetColumnByID: not found")
		}
		return Column{}, fmt.Errorf("GetColumnByID: %w", err)
	}
	return c, nil
}

// GetColumnsByBoardID returns columns for a board ordered by position.
func (db *MongoDB) GetColumnsByBoardID(ctx context.Context, boardID int) ([]Column, error) {
	columns := db.db.Collection("columns")
	cursor, err := columns.Find(
		ctx,
		bson.M{"board_id": boardID},
		options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "id", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("GetColumnsByBoardID: %w", err)
	}
	defer cursor.Close(ctx)

	var result []Column
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("GetColumnsByBoardID: %w", err)
	}

	return result, nil
}

// UpdateColumn updates a column's fields.
func (db *MongoDB) UpdateColumn(ctx context.Context, c Column) error {
	columns := db.db.Collection("columns")
	result, err := columns.UpdateOne(ctx, bson.M{"id": c.ID}, bson.M{"$set": c})
	if err != nil {
		return fmt.Errorf("UpdateColumn: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("UpdateColumn: not found")
	}
	return nil
}

// DeleteColumn deletes a column by id. Cascade deletes cards and attachments.
func (db *MongoDB) DeleteColumn(ctx context.Context, id int) error {
	columns := db.db.Collection("columns")
	result, err := columns.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("DeleteColumn: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("DeleteColumn: not found")
	}

	// Cascade delete: delete all cards and their attachments
	cards := db.db.Collection("cards")
	cardResult, err := cards.DeleteMany(ctx, bson.M{"column_id": id})
	if err != nil {
		return fmt.Errorf("DeleteColumn: cascade delete cards: %w", err)
	}

	// Delete attachments for those cards
	if cardResult.DeletedCount > 0 {
		attachments := db.db.Collection("attachments")
		_, err = attachments.DeleteMany(ctx, bson.M{"column_id": id})
		if err != nil {
			return fmt.Errorf("DeleteColumn: cascade delete attachments: %w", err)
		}
	}

	return nil
}

// CompareColumnOwner compares the owner of a column with a given user ID. Returns true if the owner matches the user ID.
func (db *MongoDB) CompareColumnOwner(ctx context.Context, id, user_id int) (bool, error) {
	var c Column
	columns := db.db.Collection("columns")
	err := columns.FindOne(ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		return false, fmt.Errorf("CompareColumnOwner: %w", err)
	}
	return c.UserID == user_id, nil
}

//
// Cards
//

// CreateCard inserts a new card and returns it with the assigned ID.
func (db *MongoDB) CreateCard(ctx context.Context, c Card) (Card, error) {
	id, err := db.getNextID(ctx, "cards")
	if err != nil {
		return Card{}, fmt.Errorf("CreateCard: getNextID: %w", err)
	}

	c.ID = id
	cards := db.db.Collection("cards")
	_, err = cards.InsertOne(ctx, c)
	if err != nil {
		return Card{}, fmt.Errorf("CreateCard: %w", err)
	}

	return c, nil
}

// GetCardByID retrieves a card by id.
func (db *MongoDB) GetCardByID(ctx context.Context, id int) (Card, error) {
	var c Card
	cards := db.db.Collection("cards")
	err := cards.FindOne(ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Card{}, fmt.Errorf("GetCardByID: not found")
		}
		return Card{}, fmt.Errorf("GetCardByID: %w", err)
	}
	return c, nil
}

// GetCardsByColumnID returns cards for a column ordered by position.
func (db *MongoDB) GetCardsByColumnID(ctx context.Context, columnID int) ([]Card, error) {
	cards := db.db.Collection("cards")
	cursor, err := cards.Find(
		ctx,
		bson.M{"column_id": columnID},
		options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "id", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("GetCardsByColumnID: %w", err)
	}
	defer cursor.Close(ctx)

	var result []Card
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("GetCardsByColumnID: %w", err)
	}

	return result, nil
}

// UpdateCard updates a card's fields.
func (db *MongoDB) UpdateCard(ctx context.Context, c Card) error {
	cards := db.db.Collection("cards")
	result, err := cards.UpdateOne(ctx, bson.M{"id": c.ID}, bson.M{"$set": c})
	if err != nil {
		return fmt.Errorf("UpdateCard: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("UpdateCard: not found")
	}
	return nil
}

// DeleteCard deletes a card by id.
func (db *MongoDB) DeleteCard(ctx context.Context, id int) error {
	cards := db.db.Collection("cards")
	result, err := cards.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("DeleteCard: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("DeleteCard: not found")
	}

	// Cascade delete attachments
	attachments := db.db.Collection("attachments")
	_, err = attachments.DeleteMany(ctx, bson.M{"card_id": id})
	if err != nil {
		return fmt.Errorf("DeleteCard: cascade delete attachments: %w", err)
	}

	return nil
}

// CompareCardOwner compares the owner of a card with a given user ID. Returns true if the owner matches the user ID.
func (db *MongoDB) CompareCardOwner(ctx context.Context, id, user_id int) (bool, error) {
	var c Card
	cards := db.db.Collection("cards")
	err := cards.FindOne(ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		return false, fmt.Errorf("CompareCardOwner: %w", err)
	}
	return c.UserID == user_id, nil
}

//
// Attachments
//

// CreateAttachment inserts a new attachment and returns it with the assigned ID.
func (db *MongoDB) CreateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	id, err := db.getNextID(ctx, "attachments")
	if err != nil {
		return Attachment{}, fmt.Errorf("CreateAttachment: getNextID: %w", err)
	}

	a.ID = id
	attachments := db.db.Collection("attachments")
	_, err = attachments.InsertOne(ctx, a)
	if err != nil {
		return Attachment{}, fmt.Errorf("CreateAttachment: %w", err)
	}

	return a, nil
}

// GetAttachmentByID retrieves an attachment by id.
func (db *MongoDB) GetAttachmentByID(ctx context.Context, id int) (Attachment, error) {
	var a Attachment
	attachments := db.db.Collection("attachments")
	err := attachments.FindOne(ctx, bson.M{"id": id}).Decode(&a)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Attachment{}, fmt.Errorf("GetAttachmentByID: not found")
		}
		return Attachment{}, fmt.Errorf("GetAttachmentByID: %w", err)
	}
	return a, nil
}

// GetAttachmentsByMultipleCardIDs retrieves all attachments belonging to the owning card ids.
func (db *MongoDB) GetAttachmentsByMultipleCardIDs(ctx context.Context, ids []int) ([]Attachment, error) {
	attachments := db.db.Collection("attachments")
	cursor, err := attachments.Find(
		ctx,
		bson.M{
			"card_id":     bson.M{"$in": ids},
			"upload_done": true,
		},
		options.Find().SetSort(bson.D{{Key: "filename", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("GetAttachmentsByMultipleCardIDs: %w", err)
	}
	defer cursor.Close(ctx)

	var result []Attachment
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("GetAttachmentsByMultipleCardIDs: %w", err)
	}

	return result, nil
}

// UpdateAttachment updates an attachment's fields.
func (db *MongoDB) UpdateAttachment(ctx context.Context, a Attachment) error {
	attachments := db.db.Collection("attachments")
	result, err := attachments.UpdateOne(ctx, bson.M{"id": a.ID}, bson.M{"$set": a})
	if err != nil {
		return fmt.Errorf("UpdateAttachment: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("UpdateAttachment: not found")
	}
	return nil
}

// DeleteAttachment deletes an attachment by id.
func (db *MongoDB) DeleteAttachment(ctx context.Context, id int) error {
	attachments := db.db.Collection("attachments")
	result, err := attachments.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("DeleteAttachment: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("DeleteAttachment: not found")
	}
	return nil
}

// CompareAttachmentOwner compares the owner of an attachment with a given user ID. Returns true if the owner matches the user ID.
func (db *MongoDB) CompareAttachmentOwner(ctx context.Context, id, user_id int) (bool, error) {
	var a Attachment
	attachments := db.db.Collection("attachments")
	err := attachments.FindOne(ctx, bson.M{"id": id}).Decode(&a)
	if err != nil {
		return false, fmt.Errorf("CompareAttachmentOwner: %w", err)
	}
	return a.UserID == user_id, nil
}
