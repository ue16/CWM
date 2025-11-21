package internal

import "context"

type Store interface {
	ListBoards(ctx context.Context) ([]Board, error)
	CreateBoard(ctx context.Context, name string) (Board, error)
	UpdateBoard(ctx context.Context, b Board) error
	GetBoardByID(ctx context.Context, id int) (Board, error)
	GetColumnsByBoardID(ctx context.Context, boardID int) ([]Column, error)
	GetCardsByColumnID(ctx context.Context, columnID int) ([]Card, error)
	DeleteBoard(ctx context.Context, id int) error
	CreateColumn(ctx context.Context, c Column) (Column, error)
	GetColumnByID(ctx context.Context, id int) (Column, error)
	UpdateColumn(ctx context.Context, c Column) error
	DeleteColumn(ctx context.Context, id int) error
	CreateCard(ctx context.Context, c Card) (Card, error)
	GetCardByID(ctx context.Context, id int) (Card, error)
	UpdateCard(ctx context.Context, c Card) error
	DeleteCard(ctx context.Context, id int) error
}
