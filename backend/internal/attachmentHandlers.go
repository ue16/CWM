package internal

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rhnvrm/simples3"
)

const (
	attachmentsBucket  = "attachments"
	presignedURLExpiry = 3600 // 1 hour
)

// prepareAttachment creates an inactive attachment record and returns a presigned PUT URL
// @Summary Prepare Attachment
// @Description Create an attachment metadata entry and get a presigned URL for uploading to S3.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param cardId path int true "Card ID"
// @Param input body internal.PrepareAttachmentRequest true "Attachment preparation payload"
// @Success 200 {object} internal.PrepareAttachmentResponse "URL and attachment ID"
// @Failure 400 {object} JSONError "Invalid input"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Card not found"
// @Failure 500 {object} JSONError "Failed to prepare attachment"
// @Router /api/v2/card/{cardId}/attachment/prepare [put]
func (s *Server) prepareAttachment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if s.s3Client == nil {
		http.Error(w, `{"error": "s3 storage not configured"}`, http.StatusInternalServerError)
		return
	}

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract card ID from URL parameters
	cardIDStr := chi.URLParam(r, "cardId")
	if cardIDStr == "" {
		http.Error(w, `{"error": "missing card ID"}`, http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid card ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Verify user owns the card
	owns, err := s.store.CompareCardOwner(r.Context(), cardID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		return
	}

	// Parse request body
	input := &PrepareAttachmentRequest{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
		return
	}

	if input.Filename == "" || input.Type == "" {
		http.Error(w, `{"error": "filename and type are required"}`, http.StatusBadRequest)
		return
	}

	// Create attachment metadata record (inactive)
	attachment := Attachment{
		CardID:     cardID,
		Filename:   input.Filename,
		Type:       input.Type,
		UploadDone: false,
		UserID:     userID,
		ObjectID:   "", // Placeholder, will be set later
	}

	createdAttachment, err := s.store.CreateAttachment(r.Context(), attachment)
	if err != nil {
		http.Error(w, `{"error": "failed to create attachment metadata"}`, http.StatusInternalServerError)
		return
	}
	attachment.ID = createdAttachment.ID

	//
	keyString := fmt.Sprintf("%d|%d|%d|%s", userID, cardID, createdAttachment.ID, input.Filename)
	hash := fnv.New64a()
	hash.Write([]byte(keyString))
	objectKey := fmt.Sprintf("%x", hash.Sum64())
	attachment.ObjectID = objectKey

	err = s.store.UpdateAttachment(r.Context(), attachment)
	if err != nil {
		http.Error(w, `{"error": "failed to create attachment metadata"}`, http.StatusInternalServerError)
		return
	}

	// Generate presigned PUT URL
	presignedURL := s.s3Client.GeneratePresignedURL(simples3.PresignedInput{
		Bucket:        attachmentsBucket,
		ObjectKey:     objectKey,
		Method:        "PUT",
		ExpirySeconds: presignedURLExpiry,
	})

	// Return response with URL and attachment ID
	response := &PrepareAttachmentResponse{
		URL: presignedURL,
		ID:  createdAttachment.ID,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// finalizeAttachment verifies the file was uploaded to S3 and marks the attachment as active
// @Summary Finalize Attachment
// @Description Verify that the file was uploaded to S3 and mark the attachment as active.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param cardId path int true "Card ID"
// @Param attachmentId path int true "Attachment ID"
// @Success 200 {object} AttachmentResponse "Attachment finalized successfully"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Attachment or card not found"
// @Failure 500 {object} JSONError "Failed to finalize attachment"
// @Router /api/v2/card/{cardId}/attachment/{attachmentId}/finalize [put]
func (s *Server) finalizeAttachment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if s.s3Client == nil {
		http.Error(w, `{"error": "s3 storage not configured"}`, http.StatusInternalServerError)
		return
	}

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract card ID from URL parameters
	cardIDStr := chi.URLParam(r, "cardId")
	if cardIDStr == "" {
		http.Error(w, `{"error": "missing card ID"}`, http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid card ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Extract attachment ID from URL parameters
	attachmentIDStr := chi.URLParam(r, "attachmentId")
	if attachmentIDStr == "" {
		http.Error(w, `{"error": "missing attachment ID"}`, http.StatusBadRequest)
		return
	}

	attachmentID, err := strconv.Atoi(attachmentIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid attachment ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Verify user owns the card
	owns, err := s.store.CompareCardOwner(r.Context(), cardID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		return
	}

	// Get attachment metadata
	attachment, err := s.store.GetAttachmentByID(r.Context(), attachmentID)
	if err != nil {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Verify attachment belongs to the specified card
	if attachment.CardID != cardID {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Verify user owns the attachment
	owns, err = s.store.CompareAttachmentOwner(r.Context(), attachmentID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Verify file exists in S3 using HEAD request
	details, err := s.s3ClientAlt.FileDetails(simples3.DetailsInput{
		Bucket:    attachmentsBucket,
		ObjectKey: attachment.ObjectID,
	})
	if err != nil {
		http.Error(w, `{"error": "file not found in storage"}`, http.StatusInternalServerError)
		return
	}

	// Update attachment metadata: mark as active and store S3 URL
	attachment.UploadDone = true
	// Parse ContentLength as int64
	if contentLen, err := strconv.ParseInt(details.ContentLength, 10, 64); err == nil {
		attachment.Size = contentLen
	}

	if err := s.store.UpdateAttachment(r.Context(), attachment); err != nil {
		http.Error(w, `{"error": "failed to finalize attachment"}`, http.StatusInternalServerError)
		return
	}

	response := &AttachmentResponse{
		ID:       attachment.ID,
		Filename: attachment.Filename,
		Type:     attachment.Type,
		Size:     attachment.Size,
		URL: s.s3Client.GeneratePresignedURL(simples3.PresignedInput{
			Bucket:        attachmentsBucket,
			ObjectKey:     attachment.ObjectID,
			Method:        "GET",
			ExpirySeconds: presignedURLExpiry,
		}),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// deleteAttachment removes attachment metadata and the object from S3
// @Summary Delete Attachment
// @Description Delete an attachment from the database and remove it from S3 storage.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from login"
// @Param cardId path int true "Card ID"
// @Param attachmentId path int true "Attachment ID"
// @Success 204 "Attachment deleted successfully"
// @Failure 401 {object} JSONError "Invalid auth token"
// @Failure 404 {object} JSONError "Attachment or card not found"
// @Failure 500 {object} JSONError "Failed to delete attachment"
// @Router /api/v2/card/{cardId}/attachment/{attachmentId} [delete]
func (s *Server) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if s.s3Client == nil {
		http.Error(w, `{"error": "s3 storage not configured"}`, http.StatusInternalServerError)
		return
	}

	userID, err := s.getUserIDFromAuthToken(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract card ID from URL parameters
	cardIDStr := chi.URLParam(r, "cardId")
	if cardIDStr == "" {
		http.Error(w, `{"error": "missing card ID"}`, http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid card ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Extract attachment ID from URL parameters
	attachmentIDStr := chi.URLParam(r, "attachmentId")
	if attachmentIDStr == "" {
		http.Error(w, `{"error": "missing attachment ID"}`, http.StatusBadRequest)
		return
	}

	attachmentID, err := strconv.Atoi(attachmentIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid attachment ID, must be an integer"}`, http.StatusBadRequest)
		return
	}

	// Verify user owns the card
	owns, err := s.store.CompareCardOwner(r.Context(), cardID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "card not found"}`, http.StatusNotFound)
		return
	}

	// Get attachment metadata
	attachment, err := s.store.GetAttachmentByID(r.Context(), attachmentID)
	if err != nil {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Verify attachment belongs to the specified card
	if attachment.CardID != cardID {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Verify user owns the attachment
	owns, err = s.store.CompareAttachmentOwner(r.Context(), attachmentID, userID)
	if err != nil || !owns {
		http.Error(w, `{"error": "attachment not found"}`, http.StatusNotFound)
		return
	}

	// Delete from S3
	objectKey := fmt.Sprintf("%d/%d/%s", cardID, attachmentID, attachment.Filename)
	err = s.s3Client.FileDelete(simples3.DeleteInput{
		Bucket:    attachmentsBucket,
		ObjectKey: objectKey,
	})
	if err != nil {
		// Log the error but don't fail the request - database cleanup is more important
		// In production, you might want to handle this differently
		log.Printf("Failed to delete attachment from S3: %v", err)
	}

	// Delete from database
	if err := s.store.DeleteAttachment(r.Context(), attachmentID); err != nil {
		http.Error(w, `{"error": "failed to delete attachment"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
