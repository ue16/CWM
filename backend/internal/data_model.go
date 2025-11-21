package internal

import (
	"fmt"
	"strings"
)

// Board represents a Kanban board.
// Matches the `boards` table in seed/kanban.sql.
type Board struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Column represents a column in a board (e.g., "To Do", "In Progress").
// Matches the `columns` table in seed/kanban.sql.
type Column struct {
	ID       int    `db:"id" json:"id"`
	BoardID  int    `db:"board_id" json:"board_id"`
	Name     string `db:"name" json:"name"`
	Position int    `db:"position" json:"position"`
}

// Card represents a task/card inside a column.
// Matches the `cards` table in seed/kanban.sql.
type Card struct {
	ID          int    `db:"id" json:"id"`
	ColumnID    int    `db:"column_id" json:"column_id"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
	Position    int    `db:"position" json:"position"`
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
 * Request structs
 */

type idRequest struct {
	ID int `json:"id"`
}

type nameRequest struct {
	Name string `json:"name"`
}
