package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Assertion to ensure DB implements the Store interface
var _ Store = (*DB)(nil)

// DB wraps a pgxpool.Pool and provides CRUD helpers for the Kanban models.
type DB struct {
	pool   *pgxpool.Pool
	cancel context.CancelFunc
}

// NewDB creates a new DB given a Postgres connection string. It retries for up
// to 5 times (max ~1 Minute) or fails.
// Example connString: "postgres://user:pass@localhost:5432/dbname"
func NewDB(connString string) (*DB, error) {
	connect := func(ctx context.Context) (*DB, error) {
		pool, err := pgxpool.New(ctx, connString)
		if err != nil {
			return nil, fmt.Errorf("pgxpool.New: %w", err)
		}

		// Verify connection
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, fmt.Errorf("ping db: %w", err)
		}

		return &DB{pool: pool}, nil
	}

	var db *DB
	var err error
	var wait time.Duration
	maxAttempts := 5
	for attempt := range maxAttempts {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err = connect(ctx)
		if err == nil {
			db.cancel = cancel
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

// Close releases the underlying pool connections.
func (db *DB) Close() {
	db.cancel()
	if db.pool != nil {
		db.pool.Close()
	}
}

//
// Users
//

// CreateUser inserts a new user and returns it with the assigned ID.
func (db *DB) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
	var id int
	var createdAt time.Time
	err := db.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`,
		username, email, passwordHash,
	).Scan(&id, &createdAt)
	if err != nil {
		return User{}, fmt.Errorf("CreateUser: %w", err)
	}
	return User{ID: id, Username: username, Email: email, PasswordHash: passwordHash, CreatedAt: createdAt}, nil
}

// GetUserByUsername retrieves a user by username.
func (db *DB) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := db.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf("GetUserByUsername: not found")
		}
		return User{}, fmt.Errorf("GetUserByUsername: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by id.
func (db *DB) GetUserByID(ctx context.Context, id int) (User, error) {
	var u User
	err := db.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf("GetUserByID: not found")
		}
		return User{}, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

// UserExists checks if a user with the given username exists.
func (db *DB) UserExists(ctx context.Context, username string) (bool, error) {
	var count int
	err := db.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE username = $1`,
		username,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("UserExists: %w", err)
	}
	return count > 0, nil
}

//
// Boards
//

// CreateBoard inserts a new board and returns it with the assigned ID.
func (db *DB) CreateBoard(ctx context.Context, name string, userID int) (Board, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO boards (name, user_id) VALUES ($1, $2) RETURNING id`,
		name, userID,
	).Scan(&id)
	if err != nil {
		return Board{}, fmt.Errorf("CreateBoard: %w", err)
	}
	return Board{ID: id, Name: name, UserID: userID}, nil
}

// GetBoardByID retrieves a board by id.
func (db *DB) GetBoardByID(ctx context.Context, id int) (Board, error) {
	var b Board
	err := db.pool.QueryRow(ctx,
		`SELECT id, name, user_id FROM boards WHERE id = $1`,
		id,
	).Scan(&b.ID, &b.Name, &b.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Board{}, fmt.Errorf("GetBoardByID: not found")
		}
		return Board{}, fmt.Errorf("GetBoardByID: %w", err)
	}
	return b, nil
}

// ListBoards returns all boards for a user.
func (db *DB) ListBoards(ctx context.Context, userID int) ([]Board, error) {
	rows, err := db.pool.Query(ctx, `SELECT id, name, user_id FROM boards WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListBoards: %w", err)
	}
	defer rows.Close()

	var boards []Board
	for rows.Next() {
		var b Board
		if err := rows.Scan(&b.ID, &b.Name, &b.UserID); err != nil {
			return nil, fmt.Errorf("ListBoards scan: %w", err)
		}
		boards = append(boards, b)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("ListBoards rows: %w", rows.Err())
	}
	return boards, nil
}

// UpdateBoard updates a board's name.
func (db *DB) UpdateBoard(ctx context.Context, b Board) error {
	ct, err := db.pool.Exec(ctx, `UPDATE boards SET name = $1 WHERE id = $2`, b.Name, b.ID)
	if err != nil {
		return fmt.Errorf("UpdateBoard: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("UpdateBoard: not found")
	}
	return nil
}

// DeleteBoard deletes a board by id. ON DELETE CASCADE will remove related columns/cards.
func (db *DB) DeleteBoard(ctx context.Context, id int) error {
	ct, err := db.pool.Exec(ctx, `DELETE FROM boards WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteBoard: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("DeleteBoard: not found")
	}
	return nil
}

// CompareBoardOwner compares the owner of a board with a given user ID. Returns true if the owner matches the user ID.
func (db *DB) CompareBoardOwner(ctx context.Context, id, user_id int) (bool, error) {
	var owner int
	row := db.pool.QueryRow(ctx, `SELECT user_id FROM boards WHERE id = $1`, id)
	err := row.Scan(&owner)
	if err != nil {
		return false, fmt.Errorf("CompareBoardOwner: %w", err)
	}
	return owner == user_id, nil
}

//
// Columns
//

// CreateColumn inserts a new column and returns it with the assigned ID.
func (db *DB) CreateColumn(ctx context.Context, c Column) (Column, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO columns (board_id, name, position, user_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		c.BoardID, c.Name, c.Position, c.UserID,
	).Scan(&id)
	if err != nil {
		return Column{}, fmt.Errorf("CreateColumn: %w", err)
	}
	c.ID = id
	return c, nil
}

// GetColumnByID retrieves a column by id.
func (db *DB) GetColumnByID(ctx context.Context, id int) (Column, error) {
	var c Column
	err := db.pool.QueryRow(ctx,
		`SELECT id, board_id, name, position FROM columns WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.BoardID, &c.Name, &c.Position)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Column{}, fmt.Errorf("GetColumnByID: not found")
		}
		return Column{}, fmt.Errorf("GetColumnByID: %w", err)
	}
	return c, nil
}

// GetColumnsByBoardID returns columns for a board ordered by position.
func (db *DB) GetColumnsByBoardID(ctx context.Context, boardID int) ([]Column, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT id, board_id, name, position FROM columns WHERE board_id = $1 ORDER BY position, id`,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetColumnsByBoardID: %w", err)
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var c Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position); err != nil {
			return nil, fmt.Errorf("GetColumnsByBoardID scan: %w", err)
		}
		cols = append(cols, c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("GetColumnsByBoardID rows: %w", rows.Err())
	}
	return cols, nil
}

// UpdateColumn updates a column's fields.
func (db *DB) UpdateColumn(ctx context.Context, c Column) error {
	ct, err := db.pool.Exec(ctx, `UPDATE columns SET board_id = $1, name = $2, position = $3 WHERE id = $4`,
		c.BoardID, c.Name, c.Position, c.ID)
	if err != nil {
		return fmt.Errorf("UpdateColumn: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("UpdateColumn: not found")
	}
	return nil
}

// DeleteColumn deletes a column by id. Cascade deletes cards.
func (db *DB) DeleteColumn(ctx context.Context, id int) error {
	ct, err := db.pool.Exec(ctx, `DELETE FROM columns WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteColumn: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("DeleteColumn: not found")
	}
	return nil
}

// CompareColumnOwner compares the owner of a column with a given user ID. Returns true if the owner matches the user ID.
func (db *DB) CompareColumnOwner(ctx context.Context, id, user_id int) (bool, error) {
	var owner int
	row := db.pool.QueryRow(ctx, `SELECT user_id FROM columns WHERE id = $1`, id)
	err := row.Scan(&owner)
	if err != nil {
		return false, fmt.Errorf("CompareColumnOwner: %w", err)
	}
	return owner == user_id, nil
}

//
// Cards
//

// CreateCard inserts a new card and returns it with the assigned ID.
func (db *DB) CreateCard(ctx context.Context, c Card) (Card, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO cards (column_id, title, description, position, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		c.ColumnID, c.Title, c.Description, c.Position, c.UserID,
	).Scan(&id)
	if err != nil {
		return Card{}, fmt.Errorf("CreateCard: %w", err)
	}
	c.ID = id
	return c, nil
}

// GetCardByID retrieves a card by id.
func (db *DB) GetCardByID(ctx context.Context, id int) (Card, error) {
	var c Card
	err := db.pool.QueryRow(ctx,
		`SELECT id, column_id, title, description, position FROM cards WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Card{}, fmt.Errorf("GetCardByID: not found")
		}
		return Card{}, fmt.Errorf("GetCardByID: %w", err)
	}
	return c, nil
}

// GetCardsByColumnID returns cards for a column ordered by position.
func (db *DB) GetCardsByColumnID(ctx context.Context, columnID int) ([]Card, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT id, column_id, title, description, position FROM cards WHERE column_id = $1 ORDER BY position, id`,
		columnID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetCardsByColumnID: %w", err)
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var c Card
		if err := rows.Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position); err != nil {
			return nil, fmt.Errorf("GetCardsByColumnID scan: %w", err)
		}
		cards = append(cards, c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("GetCardsByColumnID rows: %w", rows.Err())
	}
	return cards, nil
}

// UpdateCard updates a card's fields.
func (db *DB) UpdateCard(ctx context.Context, c Card) error {
	ct, err := db.pool.Exec(ctx,
		`UPDATE cards SET column_id = $1, title = $2, description = $3, position = $4 WHERE id = $5`,
		c.ColumnID, c.Title, c.Description, c.Position, c.ID)
	if err != nil {
		return fmt.Errorf("UpdateCard: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCard: not found")
	}
	return nil
}

// DeleteCard deletes a card by id.
func (db *DB) DeleteCard(ctx context.Context, id int) error {
	ct, err := db.pool.Exec(ctx, `DELETE FROM cards WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteCard: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("DeleteCard: not found")
	}
	return nil
}

// CompareCardOwner compares the owner of a card with a given user ID. Returns true if the owner matches the user ID.
func (db *DB) CompareCardOwner(ctx context.Context, id, user_id int) (bool, error) {
	var owner int
	row := db.pool.QueryRow(ctx, `SELECT user_id FROM cards WHERE id = $1`, id)
	err := row.Scan(&owner)
	if err != nil {
		return false, fmt.Errorf("CompareCardOwner: %w", err)
	}
	return owner == user_id, nil
}
