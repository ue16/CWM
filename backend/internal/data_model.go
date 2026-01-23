package internal

import (
	"fmt"
	"strings"
	"time"
)

// User represents a user account.
// Matches the `users` table in seed/kanban.sql.
type User struct {
	ID           int       `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// Board represents a Kanban board.
// Matches the `boards` table in seed/kanban.sql.
type Board struct {
	ID     int    `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	UserID int    `db:"user_id" json:"user_id"`
}

// Column represents a column in a board (e.g., "To Do", "In Progress").
// Matches the `columns` table in seed/kanban.sql.
type Column struct {
	ID       int    `db:"id" json:"id"`
	BoardID  int    `db:"board_id" json:"board_id"`
	Name     string `db:"name" json:"name"`
	Position int    `db:"position" json:"position"`
	UserID   int    `db:"user_id" json:"user_id"`
}

// Card represents a task/card inside a column.
// Matches the `cards` table in seed/kanban.sql.
// Attachments are exempt because we need to provide presigned URLs instead
// of object IDs to the client.
type Card struct {
	ID          int                  `db:"id" json:"id"`
	ColumnID    int                  `db:"column_id" json:"column_id"`
	Title       string               `db:"title" json:"title"`
	Description string               `db:"description" json:"description"`
	Position    int                  `db:"position" json:"position"`
	Attachments []AttachmentResponse `db:"-" json:"attachments"`
	UserID      int                  `db:"user_id" json:"user_id"`
}

// Attachment represents a file attachment info for a card.
// Matches the `attachments` table in seed/kanban.sql.
type Attachment struct {
	ID         int    `db:"id" json:"id"`
	Filename   string `db:"filename" json:"filename"`
	CardID     int    `db:"card_id" json:"card_id"`
	Size       int64  `db:"size" json:"size"`
	Type       string `db:"type" json:"type"`
	ObjectID   string `db:"object_id" json:"object_id"`
	UploadDone bool   `db:"upload_done" json:"upload_done"`
	UserID     int    `db:"user_id" json:"user_id"`
}

// String implements fmt.Stringer for Board.
func (b Board) String() string {
	return fmt.Sprintf("Board{ID:%d, Name:%q}", b.ID, b.Name)
}

// String implements fmt.Stringer for Column.
func (c Column) String() string {
	return fmt.Sprintf("Column{ID:%d, BoardID:%d, Name:%q, Position:%d}", c.ID, c.BoardID, c.Name, c.Position)
}

// String implements fmt.Stringer for Card.
// The description is truncated and newlines are collapsed for concise printing.
func (c Card) String() string {
	desc := strings.ReplaceAll(c.Description, "\n", " ")
	if len(desc) > 80 {
		desc = desc[:77] + "..."
	}
	return fmt.Sprintf("Card{ID:%d, ColumnID:%d, Title:%q, Description:%q, Position:%d}", c.ID, c.ColumnID, c.Title, desc, c.Position)
}

/*
 * Request/Response structs
 */

type IdRequest struct {
	ID int `json:"id"`
}

type NameRequest struct {
	Name string `json:"name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}

type PrepareAttachmentRequest struct {
	Filename string `json:"filename"`
	Type     string `json:"filetype"`
}

type PrepareAttachmentResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type AttachmentResponse struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Size     string `json:"size"`
	URL      string `json:"url"`
}

// For documentation

type JSONMessage struct {
	Message string `json:"message"`
}

type JSONError struct {
	Error string `json:"error"`
}
