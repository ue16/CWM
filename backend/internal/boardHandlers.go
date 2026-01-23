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
// @Param Authorization header string true "Bearer token from login"
// @Success 200 {array} internal.Board "List of boards"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to process boards"
// @Router /api/v2/boards [get]
func (s *Server) listBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
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
// @Description Fetch a board by its ID.
// @Tags Boards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param boardId path int true "Board ID"
// @Success 200 {object} Board "Board details"
// @Failure 400 {object} JSONError "Invalid board ID"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Board not found"
// @Failure 500 {object} JSONError "Failed to encode response"
// @Router /api/v2/board/{boardId} [get]
func (s *Server) getBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

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

	// Fetch the board from the database and check ownership
	board, err := s.store.GetBoardByID(r.Context(), boardID)
	if err != nil || board.UserID != userID {
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
// @Description Update a board's details.
// @Tags Boards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param User-ID header string true "User ID"
// @Param board body internal.Board true "Board object"
// @Success 200 {object} JSONMessage "Board updated successfully"
// @Failure 400 {object} JSONError "Invalid JSON body or missing fields"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to update board"
// @Router /api/v2/board/ [put]
func (s *Server) updateBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var board Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if board.ID == 0 || board.Name == "" {
		http.Error(w, `{"error": "missing required fields: id, name"}`, http.StatusBadRequest)
		return
	}

	owns, err := s.store.CompareBoardOwner(r.Context(), board.ID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
		return
	}

	board.UserID = userID
	if err := s.store.UpdateBoard(r.Context(), board); err != nil {
		http.Error(w, `{"error": "failed to update board"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "board updated successfully"}`))
}

// @Summary Create Board
// @Description Create a new board for the authenticated user.
// @Tags Boards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param User-ID header string true "User ID"
// @Param input body internal.NameRequest true "Board creation request"
// @Success 201 {object} Board "Created board"
// @Failure 400 {object} JSONError "Invalid input or missing name"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to create board"
// @Router /api/v2/board/ [post]
func (s *Server) createBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input NameRequest

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
// @Description Delete a board by its ID.
// @Tags Boards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param board body internal.IdRequest true "Board id"
// @Success 204 "Board deleted successfully"
// @Failure 400 {object} JSONError "Invalid request body"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Board not found"
// @Failure 500 {object} JSONError "Failed to delete board"
// @Router /api/v2/board/ [delete]
func (s *Server) deleteBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req IdRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	owns, err := s.store.CompareBoardOwner(r.Context(), req.ID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
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
