package internal

import "context"

// MockStore provides function fields to allow tests to configure behavior.
type MockStore struct {
	// User operations
	CreateUserFunc        func(ctx context.Context, username, email, passwordHash string) (User, error)
	GetUserByUsernameFunc func(ctx context.Context, username string) (User, error)
	GetUserByIDFunc       func(ctx context.Context, id int) (User, error)
	UserExistsFunc        func(ctx context.Context, username string) (bool, error)

	// Board operations
	ListBoardsFunc          func(ctx context.Context, userID int) ([]Board, error)
	CreateBoardFunc         func(ctx context.Context, name string, userID int) (Board, error)
	GetBoardByIDFunc        func(ctx context.Context, id int) (Board, error)
	UpdateBoardFunc         func(ctx context.Context, b Board) error
	GetColumnsByBoardIDFunc func(ctx context.Context, boardID int) ([]Column, error)
	GetCardsByColumnIDFunc  func(ctx context.Context, columnID int) ([]Card, error)
	DeleteBoardFunc         func(ctx context.Context, id int) error
	CompareBoardOwnerFunc   func(ctx context.Context, id, user_id int) (bool, error)

	// Column operations
	GetColumnByIDFunc      func(ctx context.Context, id int) (Column, error)
	CreateColumnFunc       func(ctx context.Context, c Column) (Column, error)
	UpdateColumnFunc       func(ctx context.Context, c Column) error
	DeleteColumnFunc       func(ctx context.Context, id int) error
	CompareColumnOwnerFunc func(ctx context.Context, id, user_id int) (bool, error)

	// Card operations
	GetCardByIDFunc      func(ctx context.Context, id int) (Card, error)
	CreateCardFunc       func(ctx context.Context, c Card) (Card, error)
	UpdateCardFunc       func(ctx context.Context, c Card) error
	DeleteCardFunc       func(ctx context.Context, id int) error
	CompareCardOwnerFunc func(ctx context.Context, id, user_id int) (bool, error)

	// Attachment operations
	CreateAttachmentFunc       func(ctx context.Context, a Attachment) (Attachment, error)
	GetAttachmentByIDFunc      func(ctx context.Context, id int) (Attachment, error)
	UpdateAttachmentFunc       func(ctx context.Context, a Attachment) error
	DeleteAttachmentFunc       func(ctx context.Context, id int) error
	CompareAttachmentOwnerFunc func(ctx context.Context, id, user_id int) (bool, error)
}

func (m *MockStore) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
	return m.CreateUserFunc(ctx, username, email, passwordHash)
}
func (m *MockStore) GetUserByUsername(ctx context.Context, username string) (User, error) {
	return m.GetUserByUsernameFunc(ctx, username)
}
func (m *MockStore) GetUserByID(ctx context.Context, id int) (User, error) {
	return m.GetUserByIDFunc(ctx, id)
}
func (m *MockStore) UserExists(ctx context.Context, username string) (bool, error) {
	return m.UserExistsFunc(ctx, username)
}
func (m *MockStore) ListBoards(ctx context.Context, userID int) ([]Board, error) {
	return m.ListBoardsFunc(ctx, userID)
}
func (m *MockStore) CreateBoard(ctx context.Context, name string, userID int) (Board, error) {
	return m.CreateBoardFunc(ctx, name, userID)
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
func (m *MockStore) CompareBoardOwner(ctx context.Context, id, user_id int) (bool, error) {
	return m.CompareBoardOwnerFunc(ctx, id, user_id)
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
func (m *MockStore) CompareColumnOwner(ctx context.Context, id, user_id int) (bool, error) {
	return m.CompareColumnOwnerFunc(ctx, id, user_id)
}
func (m *MockStore) CreateCard(ctx context.Context, c Card) (Card, error) {
	return m.CreateCardFunc(ctx, c)
}
func (m *MockStore) DeleteCard(ctx context.Context, id int) error {
	return m.DeleteCardFunc(ctx, id)
}
func (m *MockStore) CompareCardOwner(ctx context.Context, id, user_id int) (bool, error) {
	return m.CompareCardOwnerFunc(ctx, id, user_id)
}
func (m *MockStore) DeleteBoard(ctx context.Context, id int) error {
	return m.DeleteBoardFunc(ctx, id)
}
func (m *MockStore) CreateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	return m.CreateAttachmentFunc(ctx, a)
}
func (m *MockStore) GetAttachmentByID(ctx context.Context, id int) (Attachment, error) {
	return m.GetAttachmentByIDFunc(ctx, id)
}
func (m *MockStore) UpdateAttachment(ctx context.Context, a Attachment) error {
	return m.UpdateAttachmentFunc(ctx, a)
}
func (m *MockStore) DeleteAttachment(ctx context.Context, id int) error {
	return m.DeleteAttachmentFunc(ctx, id)
}
func (m *MockStore) CompareAttachmentOwner(ctx context.Context, id, user_id int) (bool, error) {
	return m.CompareAttachmentOwnerFunc(ctx, id, user_id)
}
