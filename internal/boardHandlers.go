package internal

import (
	"fmt"
	"log"
	"net/http"
)

func (s *Server) handleBoardDelete(w http.ResponseWriter, r *http.Request, boardID int) {
	ctx := r.Context()
	if err := s.db.DeleteBoard(ctx, boardID); err != nil {
		http.Error(w, "failed to delete board", http.StatusInternalServerError)
		log.Printf("DeleteBoard %d: %v", boardID, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleBoardCreateColumn(w http.ResponseWriter, r *http.Request, boardID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	cols, err := s.db.GetColumnsByBoardID(ctx, boardID)
	if err != nil {
		http.Error(w, "failed to load columns", http.StatusInternalServerError)
		log.Printf("GetColumnsByBoardID %d: %v", boardID, err)
		return
	}
	pos := 1
	if len(cols) > 0 {
		pos = cols[len(cols)-1].Position + 1
	}
	c := Column{BoardID: boardID, Name: name, Position: pos}
	if _, err := s.db.CreateColumn(ctx, c); err != nil {
		http.Error(w, "failed to create column", http.StatusInternalServerError)
		log.Printf("CreateColumn: %v", err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
}

func (s *Server) handleColumnEdit(w http.ResponseWriter, r *http.Request, boardID, colID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	ctx := r.Context()
	col, err := s.db.GetColumnByID(ctx, colID)
	if err != nil {
		http.Error(w, "column not found", http.StatusNotFound)
		log.Printf("GetColumnByID %d: %v", colID, err)
		return
	}
	col.Name = name
	if err := s.db.UpdateColumn(ctx, col); err != nil {
		http.Error(w, "failed to update column", http.StatusInternalServerError)
		log.Printf("UpdateColumn %d: %v", colID, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
}

func (s *Server) handleColumnMove(w http.ResponseWriter, r *http.Request, boardID, colID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	dir := r.FormValue("direction")
	ctx := r.Context()
	cols, err := s.db.GetColumnsByBoardID(ctx, boardID)
	if err != nil {
		http.Error(w, "failed to load columns", http.StatusInternalServerError)
		log.Printf("GetColumnsByBoardID %d: %v", boardID, err)
		return
	}
	idx := -1
	for i := range cols {
		if cols[i].ID == colID {
			idx = i
			break
		}
	}
	if idx == -1 {
		http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
		return
	}
	// swapping positions
	if dir == "left" && idx > 0 {
		a := cols[idx]
		b := cols[idx-1]
		a.Position, b.Position = b.Position, a.Position
		if err := s.db.UpdateColumn(ctx, a); err != nil {
			http.Error(w, "failed to move column", http.StatusInternalServerError)
			log.Printf("UpdateColumn %d: %v", a.ID, err)
			return
		}
		if err := s.db.UpdateColumn(ctx, b); err != nil {
			http.Error(w, "failed to move column", http.StatusInternalServerError)
			log.Printf("UpdateColumn %d: %v", b.ID, err)
			return
		}
	} else if dir == "right" && idx < len(cols)-1 {
		a := cols[idx]
		b := cols[idx+1]
		a.Position, b.Position = b.Position, a.Position
		if err := s.db.UpdateColumn(ctx, a); err != nil {
			http.Error(w, "failed to move column", http.StatusInternalServerError)
			log.Printf("UpdateColumn %d: %v", a.ID, err)
			return
		}
		if err := s.db.UpdateColumn(ctx, b); err != nil {
			http.Error(w, "failed to move column", http.StatusInternalServerError)
			log.Printf("UpdateColumn %d: %v", b.ID, err)
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
}

func (s *Server) handleColumnDelete(w http.ResponseWriter, r *http.Request, boardID, colID int) {
	ctx := r.Context()
	if err := s.db.DeleteColumn(ctx, colID); err != nil {
		http.Error(w, "failed to delete column", http.StatusInternalServerError)
		log.Printf("DeleteColumn %d: %v", colID, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
}

func (s *Server) handleAddCardToColumn(w http.ResponseWriter, r *http.Request, boardID, colID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	desc := r.FormValue("description")
	ctx := r.Context()
	cards, err := s.db.GetCardsByColumnID(ctx, colID)
	if err != nil {
		http.Error(w, "failed to load cards", http.StatusInternalServerError)
		log.Printf("GetCardsByColumnID %d: %v", colID, err)
		return
	}
	pos := 1
	if len(cards) > 0 {
		pos = cards[len(cards)-1].Position + 1
	}
	card := Card{ColumnID: colID, Title: title, Description: desc, Position: pos}
	if _, err := s.db.CreateCard(ctx, card); err != nil {
		http.Error(w, "failed to create card", http.StatusInternalServerError)
		log.Printf("CreateCard: %v", err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", boardID), http.StatusSeeOther)
}
