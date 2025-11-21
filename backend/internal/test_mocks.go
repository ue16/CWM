package internal

import "context"

// MockStore provides function fields to allow tests to configure behavior.
type MockStore struct {
	ListBoardsFunc          func(ctx context.Context) ([]Board, error)
	CreateBoardFunc         func(ctx context.Context, name string) (Board, error)
	GetBoardByIDFunc        func(ctx context.Context, id int) (Board, error)
	UpdateBoardFunc         func(ctx context.Context, b Board) error
	GetColumnsByBoardIDFunc func(ctx context.Context, boardID int) ([]Column, error)
	GetCardsByColumnIDFunc  func(ctx context.Context, columnID int) ([]Card, error)
	GetColumnByIDFunc       func(ctx context.Context, id int) (Column, error)
	GetCardByIDFunc         func(ctx context.Context, id int) (Card, error)
	UpdateCardFunc          func(ctx context.Context, c Card) error
	CreateColumnFunc        func(ctx context.Context, c Column) (Column, error)
	UpdateColumnFunc        func(ctx context.Context, c Column) error
	DeleteColumnFunc        func(ctx context.Context, id int) error
	CreateCardFunc          func(ctx context.Context, c Card) (Card, error)
	DeleteCardFunc          func(ctx context.Context, id int) error
	DeleteBoardFunc         func(ctx context.Context, id int) error
}

func (m *MockStore) ListBoards(ctx context.Context) ([]Board, error) {
	return m.ListBoardsFunc(ctx)
}
func (m *MockStore) CreateBoard(ctx context.Context, name string) (Board, error) {
	return m.CreateBoardFunc(ctx, name)
}
func (m *MockStore) UpdateBoard(ctx context.Context, b Board) error {
	return m.UpdateBoardFunc(ctx, b)
}
func (m *MockStore) GetBoardByID(ctx context.Context, id int) (Board, error) {
	return m.GetBoardByIDFunc(ctx, id)
}
func (m *MockStore) GetColumnsByBoardID(ctx context.Context, boardID int) ([]Column, error) {
	return m.GetColumnsByBoardIDFunc(ctx, boardID)
}
func (m *MockStore) GetCardsByColumnID(ctx context.Context, columnID int) ([]Card, error) {
	return m.GetCardsByColumnIDFunc(ctx, columnID)
}
func (m *MockStore) GetColumnByID(ctx context.Context, id int) (Column, error) {
	return m.GetColumnByIDFunc(ctx, id)
}
func (m *MockStore) GetCardByID(ctx context.Context, id int) (Card, error) {
	return m.GetCardByIDFunc(ctx, id)
}
func (m *MockStore) UpdateCard(ctx context.Context, c Card) error {
	return m.UpdateCardFunc(ctx, c)
}
func (m *MockStore) CreateColumn(ctx context.Context, c Column) (Column, error) {
	return m.CreateColumnFunc(ctx, c)
}
func (m *MockStore) UpdateColumn(ctx context.Context, c Column) error {
	return m.UpdateColumnFunc(ctx, c)
}
func (m *MockStore) DeleteColumn(ctx context.Context, id int) error {
	return m.DeleteColumnFunc(ctx, id)
}
func (m *MockStore) CreateCard(ctx context.Context, c Card) (Card, error) {
	return m.CreateCardFunc(ctx, c)
}
func (m *MockStore) DeleteCard(ctx context.Context, id int) error {
	return m.DeleteCardFunc(ctx, id)
}
func (m *MockStore) DeleteBoard(ctx context.Context, id int) error {
	return m.DeleteBoardFunc(ctx, id)
}
