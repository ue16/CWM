package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// login handles user login and returns a token.
// @Summary Login user
// @Description Authenticate a user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param loginReq body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/v2/auth/login [post]
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Generate token (simple JWT-like token with user ID and expiration)
	token := generateToken(user.ID)

	response := LoginResponse{
		Token: token,
		User: User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// register handles user registration.
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param registerReq body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request or user exists"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/v2/auth/register [post]
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"username, email, and password required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, `{"error":"password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	exists, err := s.store.UserExists(r.Context(), req.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, `{"error":"username already exists"}`, http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := s.store.CreateUser(r.Context(), req.Username, req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	// Generate token
	token := generateToken(user.ID)

	response := AuthResponse{
		Success: true,
		Message: "User created successfully",
		Token:   token,
		User: &User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// verify checks if a token is valid and returns the user.
// @Summary Verify authentication token
// @Description Verify if the provided token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Param token header string true "Auth token"
// @Success 200 {object} User "Token is valid"
// @Failure 401 {object} map[string]string "Invalid token"
// @Router /api/v2/auth/verify [get]
func (s *Server) verify(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, `{"error":"token required"}`, http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userID := verifyToken(token)
	if userID == 0 {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// Simple token generation and verification (in production, use proper JWT)
func generateToken(userID int) string {
	// For this simple implementation, we use a basic format
	// In production, you should use proper JWT tokens
	return fmt.Sprintf("%d:%d", userID, time.Now().Add(24*time.Hour).Unix())
}

func verifyToken(token string) int {
	// Parse the simple token format
	var userID int
	var expiry int64
	_, err := fmt.Sscanf(token, "%d:%d", &userID, &expiry)
	if err != nil {
		return 0
	}

	// Check if token has expired
	if time.Now().Unix() > expiry {
		return 0
	}

	return userID
}

// getUserIDFromAuthToken extracts and verifies the user ID from the Authorization header
// This should be used in all protected routes to ensure user authentication
func getUserIDFromAuthToken(r *http.Request) (int, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return 0, fmt.Errorf("missing authorization token")
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userID := verifyToken(token)
	if userID == 0 {
		return 0, fmt.Errorf("invalid or expired token")
	}

	return userID, nil
}

// isBoardOwner checks if the given user owns the specified board
func isBoardOwner(ctx context.Context, store Store, userID, boardID int) (bool, error) {
	board, err := store.GetBoardByID(ctx, boardID)
	if err != nil {
		return false, err
	}
	return board.UserID == userID, nil
}

// isColumnBelongsToBoardOwner checks if the user owns the board that contains this column
// This verifies the column belongs to a board owned by the user
func isColumnBelongsToBoardOwner(ctx context.Context, store Store, userID, columnID int) (bool, error) {
	column, err := store.GetColumnByID(ctx, columnID)
	if err != nil {
		return false, err
	}
	return isBoardOwner(ctx, store, userID, column.BoardID)
}

// isCardBelongsToBoardOwner checks if the user owns the board that contains this card
// This verifies the card belongs to a column in a board owned by the user
func isCardBelongsToBoardOwner(ctx context.Context, store Store, userID, cardID int) (bool, error) {
	card, err := store.GetCardByID(ctx, cardID)
	if err != nil {
		return false, err
	}
	column, err := store.GetColumnByID(ctx, card.ColumnID)
	if err != nil {
		return false, err
	}
	return isBoardOwner(ctx, store, userID, column.BoardID)
}
