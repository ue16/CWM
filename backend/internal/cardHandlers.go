package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// @Summary Get Cards
// @Description Fetch all cards for a specific column.
// @Tags Cards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param columnId path int true "Column ID"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Router /api/v2/card/{columnId} [get]
func (s *Server) getCards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract column ID from URL parameters
	columnIDStr := chi.URLParam(r, "columnId")
	if columnIDStr == "" {
		http.Error(w, `{"error": "missing column ID"}`, http.StatusBadRequest)
		return
	}

	columnID, err := strconv.Atoi(columnIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid column ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	owns, err := s.store.CompareColumnOwner(r.Context(), columnID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "column not found"}`, http.StatusNotFound)
		return
	}

	// Fetch cards for the specified column
	cards, err := s.store.GetCardsByColumnID(r.Context(), columnID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch cards"}`, http.StatusInternalServerError)
		return
	}

	var ids []int
	for _, card := range cards {
		ids = append(ids, card.ID)
	}
	attachments, err := s.store.GetAttachmentsByMultipleCardIDs(r.Context(), ids)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch attachments"}`, http.StatusInternalServerError)
		log.Printf("failed to fetch attachments: %v", err)
		return
	}
	tmpAttachments := map[int][]AttachmentResponse{}
	for _, attachment := range attachments {
		if tmpAttachments[attachment.CardID] == nil {
			tmpAttachments[attachment.CardID] = []AttachmentResponse{}
		}
		tmpAttachments[attachment.CardID] = append(tmpAttachments[attachment.CardID], AttachmentResponse{
			ID:       attachment.ID,
			Filename: attachment.Filename,
			Type:     attachment.Type,
			Size:     humanizeBytes(attachment.Size),
		})
	}
	for idx := range cards {
		cards[idx].Attachments = tmpAttachments[cards[idx].ID]
	}

	// Respond with the cards data
	if err := json.NewEncoder(w).Encode(cards); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// @Summary Create Card
// @Description Create a new card in a specific column.
// @Tags Cards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param card body object true "Card creation payload"
// @Param card.column_id body int true "Column ID"
// @Param card.title body string true "Card title"
// @Param card.description body string false "Card description"
// @Success 201 {object} Card
// @Failure 400 {object} JSONError "Invalid input or missing required fields"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to create card"
// @Router /api/v2/card/ [post]
func (s *Server) createCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input struct {
		ColumnID    int    `json:"column_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if input.ColumnID == 0 || input.Title == "" {
		http.Error(w, `{"error": "column_id and title are required"}`, http.StatusBadRequest)
		return
	}

	owns, err := s.store.CompareColumnOwner(r.Context(), input.ColumnID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "column not found"}`, http.StatusNotFound)
		return
	}

	// Fetch existing cards to determine position
	cards, err := s.store.GetCardsByColumnID(r.Context(), input.ColumnID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch cards"}`, http.StatusInternalServerError)
		return
	}

	position := 1
	if len(cards) > 0 {
		position = cards[len(cards)-1].Position + 1
	}

	// Create the card
	card := Card{
		ColumnID:    input.ColumnID,
		Title:       input.Title,
		Description: input.Description,
		Position:    position,
		UserID:      userID,
	}

	createdCard, err := s.store.CreateCard(r.Context(), card)
	if err != nil {
		http.Error(w, `{"error": "failed to create card"}`, http.StatusInternalServerError)
		return
	}

	// Respond with the created card
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdCard); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// @Summary Update Card
// @Description Update a card's details.
// @Tags Cards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param card body internal.Card true "Card object"
// @Success 200 {object} JSONMessage "Card updated successfully"
// @Failure 400 {object} JSONError "Invalid JSON body or missing fields"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 500 {object} JSONError "Failed to update card"
// @Router /api/v2/card/ [put]
func (s *Server) updateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var card Card
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if card.ID == 0 || card.Title == "" || card.Description == "" {
		http.Error(w, `{"error": "missing required fields: id, title, description"}`, http.StatusBadRequest)
		return
	}

	owns, err := s.store.CompareCardOwner(r.Context(), card.ID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		return
	}

	card.UserID = userID
	if err := s.store.UpdateCard(r.Context(), card); err != nil {
		http.Error(w, `{"error": "failed to update card"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "card updated successfully"}`))
}

// @Summary Delete Card
// @Description Delete a card by its ID.
// @Tags Cards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param card body internal.IdRequest true "Card id"
// @Success 204 "Card deleted successfully"
// @Failure 400 {object} JSONError "Invalid request body"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Card not found"
// @Failure 500 {object} JSONError "Failed to delete card"
// @Router /api/v2/card/ [delete]
func (s *Server) deleteCard(w http.ResponseWriter, r *http.Request) {
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

	owns, err := s.store.CompareCardOwner(r.Context(), req.ID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		return
	}

	if err := s.store.DeleteCard(r.Context(), req.ID); err != nil {
		if err.Error() == "DeleteCard: not found" {
			http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "failed to delete card"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
