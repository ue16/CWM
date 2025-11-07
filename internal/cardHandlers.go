package internal

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (s *Server) handleCardEdit(w http.ResponseWriter, r *http.Request, id int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	desc := r.FormValue("description")

	ctx := r.Context()
	card, err := s.store.GetCardByID(ctx, id)
	if err != nil {
		http.Error(w, "card not found", http.StatusNotFound)
		log.Printf("GetCardByID %d: %v", id, err)
		return
	}
	card.Title = title
	card.Description = desc
	if err := s.store.UpdateCard(ctx, card); err != nil {
		http.Error(w, "failed to update card", http.StatusInternalServerError)
		log.Printf("UpdateCard %d: %v", id, err)
		return
	}
	col, err := s.store.GetColumnByID(ctx, card.ColumnID)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
}

func (s *Server) handleCardMove(w http.ResponseWriter, r *http.Request, id int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	// If a target column is specified, move directly there.
	if tgt := r.FormValue("target_column_id"); tgt != "" {
		targetID, err := strconv.Atoi(tgt)
		if err != nil {
			http.Error(w, "invalid target column", http.StatusBadRequest)
			return
		}
		// compute new position as end of target column
		cards, err := s.store.GetCardsByColumnID(ctx, targetID)
		if err != nil {
			http.Error(w, "failed to load cards", http.StatusInternalServerError)
			log.Printf("GetCardsByColumnID %d: %v", targetID, err)
			return
		}
		pos := 1
		if len(cards) > 0 {
			pos = cards[len(cards)-1].Position + 1
		}

		card, err := s.store.GetCardByID(ctx, id)
		if err != nil {
			http.Error(w, "card not found", http.StatusNotFound)
			log.Printf("GetCardByID %d: %v", id, err)
			return
		}
		card.ColumnID = targetID
		card.Position = pos
		if err := s.store.UpdateCard(ctx, card); err != nil {
			http.Error(w, "failed to move card", http.StatusInternalServerError)
			log.Printf("UpdateCard %d: %v", id, err)
			return
		}
		col, err := s.store.GetColumnByID(ctx, targetID)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
		return
	}

	// Otherwise handle directional move (left/right)
	dir := r.FormValue("direction")
	card, err := s.store.GetCardByID(ctx, id)
	if err != nil {
		http.Error(w, "card not found", http.StatusNotFound)
		log.Printf("GetCardByID %d: %v", id, err)
		return
	}
	col, err := s.store.GetColumnByID(ctx, card.ColumnID)
	if err != nil {
		http.Error(w, "failed to load column", http.StatusInternalServerError)
		log.Printf("GetColumnByID %d: %v", card.ColumnID, err)
		return
	}
	cols, err := s.store.GetColumnsByBoardID(ctx, col.BoardID)
	if err != nil {
		http.Error(w, "failed to load columns", http.StatusInternalServerError)
		log.Printf("GetColumnsByBoardID %d: %v", col.BoardID, err)
		return
	}
	idx := -1
	for i := range cols {
		if cols[i].ID == col.ID {
			idx = i
			break
		}
	}
	if idx == -1 {
		http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
		return
	}
	var destCol Column
	if dir == "left" && idx > 0 {
		destCol = cols[idx-1]
	} else if dir == "right" && idx < len(cols)-1 {
		destCol = cols[idx+1]
	} else {
		// nothing to do
		http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
		return
	}
	// append to end of dest column
	destCards, err := s.store.GetCardsByColumnID(ctx, destCol.ID)
	if err != nil {
		http.Error(w, "failed to load cards", http.StatusInternalServerError)
		log.Printf("GetCardsByColumnID %d: %v", destCol.ID, err)
		return
	}
	pos := 1
	if len(destCards) > 0 {
		pos = destCards[len(destCards)-1].Position + 1
	}
	card.ColumnID = destCol.ID
	card.Position = pos
	if err := s.store.UpdateCard(ctx, card); err != nil {
		http.Error(w, "failed to move card", http.StatusInternalServerError)
		log.Printf("UpdateCard %d: %v", id, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
}

func (s *Server) handleCardDelete(w http.ResponseWriter, r *http.Request, id int) {
	ctx := r.Context()
	// get card to discover board
	card, err := s.store.GetCardByID(ctx, id)
	if err != nil {
		http.Error(w, "card not found", http.StatusNotFound)
		log.Printf("GetCardByID %d: %v", id, err)
		return
	}
	col, err := s.store.GetColumnByID(ctx, card.ColumnID)
	if err != nil {
		http.Error(w, "failed to load column", http.StatusInternalServerError)
		log.Printf("GetColumnByID %d: %v", card.ColumnID, err)
		return
	}
	if err := s.store.DeleteCard(ctx, id); err != nil {
		http.Error(w, "failed to delete card", http.StatusInternalServerError)
		log.Printf("DeleteCard %d: %v", id, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/%d", col.BoardID), http.StatusSeeOther)
}
