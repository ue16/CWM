package internal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// @Summary List all boards for a user
// @Description Lists all boards belonging to the authenticated user.
// @Tags All-Boards
// @Accept json
// @Produce json
// @Param User-ID header string true "User ID"
// @Success 200 {array} internal.Board "List of boards"
// @Failure 400 {object} map[string]string "Invalid or missing user ID"
// @Failure 500 {object} map[string]string "Failed to list boards"
// @Router /api/v2/boards [get]
func (s *Server) listBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	boards, err := s.store.ListBoards(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "failed to list boards"}`, http.StatusInternalServerError)
		return
	}

	if boards == nil {
		boards = []Board{}
	}

	if boardsJSON, err := json.Marshal(boards); err == nil {
		w.Write(boardsJSON)
	} else {
		http.Error(w, `{"error": "failed to marshal boards"}`, http.StatusInternalServerError)
		return
	}
}

// @Summary Get Board
// @Description Fetch a board by its ID
// @Tags Boards
// @Accept json
// @Produce json
// @Param boardId path int true "Board ID"
// @Success 200 {object} Board "Board details"
// @Failure 400 {object} map[string]string "Invalid board ID"
// @Failure 404 {object} map[string]string "Board not found"
// @Failure 500 {object} map[string]string "Failed to encode response"
// @Router /api/v2/board/{boardId} [get]
func (s *Server) getBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Extract board ID from URL parameters
	boardIDStr := chi.URLParam(r, "boardId")
	if boardIDStr == "" {
		http.Error(w, `{"error": "missing board ID"}`, http.StatusBadRequest)
		return
	}

	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid board ID"}`, http.StatusBadRequest)
		return
	}

	// Fetch the board from the database
	board, err := s.store.GetBoardByID(r.Context(), boardID)
	if err != nil {
		http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
		return
	}

	// Respond with the board data
	if err := json.NewEncoder(w).Encode(board); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// @Summary Update Board
// @Description Update a board's details
// @Tags Boards
// @Accept json
// @Produce json
// @Param board body internal.Board true "Board object"
// @Success 200 {object} map[string]string "Board updated successfully"
// @Failure 400 {object} map[string]string "Invalid JSON body or missing fields"
// @Failure 500 {object} map[string]string "Failed to update board"
// @Router /api/v2/board/ [put]
func (s *Server) updateBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var board Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		http.Error(w, "{\"err\": \"invalid JSON body\"}", http.StatusBadRequest)
		return
	}

	if board.ID == 0 || board.Name == "" {
		http.Error(w, "{\"err\": \"missing required fields: id, name\"}", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateBoard(r.Context(), board); err != nil {
		http.Error(w, "{\"err\": \"failed to update board\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"msg\": \"board updated successfully\"}"))
}

// @Summary Create Board
// @Description Create a new board for the authenticated user
// @Tags Boards
// @Accept json
// @Produce json
// @Param User-ID header string true "User ID"
// @Param input body internal.nameRequest true "Board creation request"
// @Success 201 {object} Board "Created board"
// @Failure 400 {object} map[string]string "Invalid input or missing name"
// @Failure 500 {object} map[string]string "Failed to create board"
// @Router /api/v2/board/ [post]
func (s *Server) createBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input nameRequest

	// Decode the JSON body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if input.Name == "" {
		http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
		return
	}

	// Create the board
	board, err := s.store.CreateBoard(r.Context(), input.Name, userID)
	if err != nil {
		http.Error(w, `{"error": "failed to create board"}`, http.StatusInternalServerError)
		return
	}

	// Respond with the created board
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(board); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// @Summary Delete Board
// @Description Delete a board by its ID
// @Tags Boards
// @Accept json
// @Produce json
// @Param board body internal.idRequest true "Board id"
// @Success 204 "Board deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 404 {object} map[string]string "Board not found"
// @Failure 500 {object} map[string]string "Failed to delete board"
// @Router /api/v2/board/ [delete]
func (s *Server) deleteBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req idRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := s.store.DeleteBoard(r.Context(), req.ID); err != nil {
		if err.Error() == "DeleteBoard: not found" {
			http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "failed to delete board"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
