package internal

import "context"

type Store interface {
	// User operations
	CreateUser(ctx context.Context, username, email, passwordHash string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	UserExists(ctx context.Context, username string) (bool, error)

	// Board operations
	ListBoards(ctx context.Context, userID int) ([]Board, error)
	CreateBoard(ctx context.Context, name string, userID int) (Board, error)
	UpdateBoard(ctx context.Context, b Board) error
	GetBoardByID(ctx context.Context, id int) (Board, error)
	GetColumnsByBoardID(ctx context.Context, boardID int) ([]Column, error)
	GetCardsByColumnID(ctx context.Context, columnID int) ([]Card, error)
	DeleteBoard(ctx context.Context, id int) error

	// Column operations
	CreateColumn(ctx context.Context, c Column) (Column, error)
	GetColumnByID(ctx context.Context, id int) (Column, error)
	UpdateColumn(ctx context.Context, c Column) error
	DeleteColumn(ctx context.Context, id int) error

	// Card operations
	CreateCard(ctx context.Context, c Card) (Card, error)
	GetCardByID(ctx context.Context, id int) (Card, error)
	UpdateCard(ctx context.Context, c Card) error
	DeleteCard(ctx context.Context, id int) error
}
