package internal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// @Summary Get Columns
// @Description Fetch all columns for a specific board.
// @Tags Columns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param boardId path int true "Board ID"
// @Success 200 {array} Column
// @Failure 400 {object} JSONError "Invalid board ID"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to fetch columns"
// @Router /api/v2/column/{boardId} [get]
func (s *Server) getColumns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
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
		http.Error(w, `{"error": "invalid board ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Check if user owns this board
	owns, err := isBoardOwner(r.Context(), s.store, userID, boardID)
	if err != nil || !owns {
		http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
		return
	}

	// Fetch columns for the specified board
	columns, err := s.store.GetColumnsByBoardID(r.Context(), boardID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch columns"}`, http.StatusInternalServerError)
		return
	}

	// Respond with the columns data
	if err := json.NewEncoder(w).Encode(columns); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// @Summary Create Column
// @Description Create a new column in a specific board.
// @Tags Columns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param input body object true "Column creation payload"
// @Param input.board_id body int true "Board ID"
// @Param input.name body string true "Column name"
// @Success 201 {object} Column
// @Failure 400 {object} JSONError "Invalid input or missing fields"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to create column"
// @Router /api/v2/column/ [post]
func (s *Server) createColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input struct {
		BoardID int    `json:"board_id"`
		Name    string `json:"name"`
	}

	// Decode the JSON body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if input.BoardID == 0 || input.Name == "" {
		http.Error(w, `{"error": "board_id and name are required"}`, http.StatusBadRequest)
		return
	}

	// Check if user owns this board
	owns, err := isBoardOwner(r.Context(), s.store, userID, input.BoardID)
	if err != nil || !owns {
		http.Error(w, `{"error": "board not found"}`, http.StatusNotFound)
		return
	}

	// Fetch existing columns to determine position
	columns, err := s.store.GetColumnsByBoardID(r.Context(), input.BoardID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch columns"}`, http.StatusInternalServerError)
		return
	}

	position := 1
	if len(columns) > 0 {
		position = columns[len(columns)-1].Position + 1
	}

	// Create the column
	column := Column{
		BoardID:  input.BoardID,
		Name:     input.Name,
		Position: position,
	}

	createdColumn, err := s.store.CreateColumn(r.Context(), column)
	if err != nil {
		http.Error(w, `{"error": "failed to create column"}`, http.StatusInternalServerError)
		return
	}

	// Respond with the created column
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdColumn); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// @Summary Update Column
// @Description Update a column's details.
// @Tags Columns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param input body internal.Column true "Column object"
// @Success 200 {object} JSONMessage "Column updated successfully"
// @Failure 400 {object} JSONError "Invalid JSON body or missing fields"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to update column"
// @Router /api/v2/column/ [put]
func (s *Server) updateColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, "{\"error\": \"unauthorized\"}", http.StatusUnauthorized)
		return
	}

	var column Column
	if err := json.NewDecoder(r.Body).Decode(&column); err != nil {
		http.Error(w, "{\"error\": \"invalid JSON body\"}", http.StatusBadRequest)
		return
	}

	if column.ID == 0 || column.Name == "" {
		http.Error(w, "{\"error\": \"missing required fields: id, name\"}", http.StatusBadRequest)
		return
	}

	// Check if user owns the board that contains this column
	owns, err := isColumnBelongsToBoardOwner(r.Context(), s.store, userID, column.ID)
	if err != nil || !owns {
		http.Error(w, "{\"error\": \"column not found\"}", http.StatusNotFound)
		return
	}

	if err := s.store.UpdateColumn(r.Context(), column); err != nil {
		http.Error(w, "{\"error\": \"failed to update column\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"column updated successfully\"}"))
}

// @Summary Delete Column
// @Description Delete a column by its ID.
// @Tags Columns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param input body internal.idRequest true "Column ID"
// @Success 204 "Column deleted successfully"
// @Failure 400 {object} JSONError "Invalid request body"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Column not found"
// @Failure 500 {object} JSONError "Failed to delete column"
// @Router /api/v2/column/ [delete]
func (s *Server) deleteColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req idRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Check if user owns the board that contains this column
	owns, err := isColumnBelongsToBoardOwner(r.Context(), s.store, userID, req.ID)
	if err != nil || !owns {
		http.Error(w, `{"error": "column not found"}`, http.StatusNotFound)
		return
	}

	if err := s.store.DeleteColumn(r.Context(), req.ID); err != nil {
		if err.Error() == "DeleteColumn: not found" {
			http.Error(w, `{"error": "column not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "failed to delete column"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
