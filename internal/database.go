package internal

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgxpool.Pool and provides CRUD helpers for the Kanban models.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB creates a new DB given a Postgres connection string.
// Example connString: "postgres://user:pass@localhost:5432/dbname"
func NewDB(ctx context.Context, connString string) (*DB, error) {
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

// Close releases the underlying pool connections.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

//
// Boards
//

// CreateBoard inserts a new board and returns it with the assigned ID.
func (db *DB) CreateBoard(ctx context.Context, name string) (Board, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO boards (name) VALUES ($1) RETURNING id`,
		name,
	).Scan(&id)
	if err != nil {
		return Board{}, fmt.Errorf("CreateBoard: %w", err)
	}
	return Board{ID: id, Name: name}, nil
}

// GetBoardByID retrieves a board by id.
func (db *DB) GetBoardByID(ctx context.Context, id int) (Board, error) {
	var b Board
	err := db.pool.QueryRow(ctx,
		`SELECT id, name FROM boards WHERE id = $1`,
		id,
	).Scan(&b.ID, &b.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Board{}, fmt.Errorf("GetBoardByID: not found")
		}
		return Board{}, fmt.Errorf("GetBoardByID: %w", err)
	}
	return b, nil
}

// ListBoards returns all boards.
func (db *DB) ListBoards(ctx context.Context) ([]Board, error) {
	rows, err := db.pool.Query(ctx, `SELECT id, name FROM boards ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("ListBoards: %w", err)
	}
	defer rows.Close()

	var boards []Board
	for rows.Next() {
		var b Board
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
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

//
// Columns
//

// CreateColumn inserts a new column and returns it with the assigned ID.
func (db *DB) CreateColumn(ctx context.Context, c Column) (Column, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO columns (board_id, name, position) VALUES ($1, $2, $3) RETURNING id`,
		c.BoardID, c.Name, c.Position,
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

//
// Cards
//

// CreateCard inserts a new card and returns it with the assigned ID.
func (db *DB) CreateCard(ctx context.Context, c Card) (Card, error) {
	var id int
	err := db.pool.QueryRow(ctx,
		`INSERT INTO cards (column_id, title, description, position) VALUES ($1, $2, $3, $4) RETURNING id`,
		c.ColumnID, c.Title, c.Description, c.Position,
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
