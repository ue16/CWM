package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		userExists    bool
		expectStatus  int
		expectSuccess bool
		expectMessage string
	}{
		{
			name:          "successful registration",
			body:          `{"username":"newuser","email":"newuser@example.com","password":"password123"}`,
			userExists:    false,
			expectStatus:  http.StatusCreated,
			expectSuccess: true,
			expectMessage: "User created successfully",
		},
		{
			name:          "user already exists",
			body:          `{"username":"existing","email":"existing@example.com","password":"password123"}`,
			userExists:    true,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "username already exists",
		},
		{
			name:          "missing username",
			body:          `{"email":"test@example.com","password":"password123"}`,
			userExists:    false,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "username, email, and password required",
		},
		{
			name:          "missing email",
			body:          `{"username":"testuser","password":"password123"}`,
			userExists:    false,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "username, email, and password required",
		},
		{
			name:          "missing password",
			body:          `{"username":"testuser","email":"test@example.com"}`,
			userExists:    false,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "username, email, and password required",
		},
		{
			name:          "password too short",
			body:          `{"username":"testuser","email":"test@example.com","password":"short"}`,
			userExists:    false,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "password must be at least 6 characters",
		},
		{
			name:          "invalid json",
			body:          `{invalid json}`,
			userExists:    false,
			expectStatus:  http.StatusBadRequest,
			expectSuccess: false,
			expectMessage: "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			user := User{
				ID:           1,
				Username:     "newuser",
				Email:        "newuser@example.com",
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
			}

			mock := &MockStore{
				UserExistsFunc: func(ctx context.Context, username string) (bool, error) {
					return tt.userExists, nil
				},
				CreateUserFunc: func(ctx context.Context, username, email, passwordHash string) (User, error) {
					return user, nil
				},
			}
			s := &Server{store: mock}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			s.register(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if tt.expectSuccess {
				var authResp AuthResponse
				if err := json.Unmarshal(body, &authResp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !authResp.Success {
					t.Fatalf("expected success to be true")
				}
				if authResp.Message != tt.expectMessage {
					t.Fatalf("expected message %q, got %q", tt.expectMessage, authResp.Message)
				}
				if authResp.Token == "" {
					t.Fatalf("expected token to be set")
				}
				if authResp.User == nil {
					t.Fatalf("expected user to be set")
				}
				if authResp.User.Username != "newuser" {
					t.Fatalf("expected username %q, got %q", "newuser", authResp.User.Username)
				}
			} else {
				if !strings.Contains(bodyStr, tt.expectMessage) {
					t.Fatalf("expected error message %q in response, got %s", tt.expectMessage, bodyStr)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		body          string
		userFound     bool
		expectStatus  int
		expectToken   bool
		expectMessage string
	}{
		{
			name:         "successful login",
			body:         `{"username":"testuser","password":"correctpassword"}`,
			userFound:    true,
			expectStatus: http.StatusOK,
			expectToken:  true,
		},
		{
			name:          "invalid credentials - wrong password",
			body:          `{"username":"testuser","password":"wrongpassword"}`,
			userFound:     true,
			expectStatus:  http.StatusUnauthorized,
			expectToken:   false,
			expectMessage: "invalid credentials",
		},
		{
			name:          "user not found",
			body:          `{"username":"nonexistent","password":"somepassword"}`,
			userFound:     false,
			expectStatus:  http.StatusUnauthorized,
			expectToken:   false,
			expectMessage: "invalid credentials",
		},
		{
			name:          "missing username",
			body:          `{"password":"password123"}`,
			userFound:     false,
			expectStatus:  http.StatusBadRequest,
			expectToken:   false,
			expectMessage: "username and password required",
		},
		{
			name:          "missing password",
			body:          `{"username":"testuser"}`,
			userFound:     false,
			expectStatus:  http.StatusBadRequest,
			expectToken:   false,
			expectMessage: "username and password required",
		},
		{
			name:          "invalid json",
			body:          `{invalid json}`,
			userFound:     false,
			expectStatus:  http.StatusBadRequest,
			expectToken:   false,
			expectMessage: "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStore{
				GetUserByUsernameFunc: func(ctx context.Context, username string) (User, error) {
					if tt.userFound {
						return user, nil
					}
					return User{}, errors.New("user not found")
				},
			}
			s := &Server{store: mock}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			s.login(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if tt.expectToken {
				var loginResp LoginResponse
				if err := json.Unmarshal(body, &loginResp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if loginResp.Token == "" {
					t.Fatalf("expected token to be set")
				}
				if loginResp.User.Username != "testuser" {
					t.Fatalf("expected username %q, got %q", "testuser", loginResp.User.Username)
				}
			} else {
				if !strings.Contains(bodyStr, tt.expectMessage) {
					t.Fatalf("expected error message %q in response, got %s", tt.expectMessage, bodyStr)
				}
			}
		})
	}
}

func TestVerify(t *testing.T) {
	secret := "secret"
	user := User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	validToken := generateToken(user.ID, secret, 24)

	tests := []struct {
		name          string
		token         string
		userFound     bool
		expectStatus  int
		expectUser    bool
		expectMessage string
	}{
		{
			name:         "valid token",
			token:        "Bearer " + validToken,
			userFound:    true,
			expectStatus: http.StatusOK,
			expectUser:   true,
		},
		{
			name:          "invalid token format",
			token:         "Bearer invalid-token",
			userFound:     false,
			expectStatus:  http.StatusUnauthorized,
			expectUser:    false,
			expectMessage: "invalid token",
		},
		{
			name:          "expired token",
			token:         "Bearer " + generateToken(1, secret, -1),
			userFound:     false,
			expectStatus:  http.StatusUnauthorized,
			expectUser:    false,
			expectMessage: "invalid token",
		},
		{
			name:          "missing token",
			token:         "",
			userFound:     false,
			expectStatus:  http.StatusUnauthorized,
			expectUser:    false,
			expectMessage: "token required",
		},
		{
			name:          "user not found with valid token",
			token:         "Bearer " + validToken,
			userFound:     false,
			expectStatus:  http.StatusUnauthorized,
			expectUser:    false,
			expectMessage: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStore{
				GetUserByIDFunc: func(ctx context.Context, id int) (User, error) {
					if tt.userFound {
						return user, nil
					}
					return User{}, errors.New("user not found")
				},
			}
			s := &Server{secret: secret, store: mock}

			req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			rr := httptest.NewRecorder()
			s.verify(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if tt.expectUser {
				var verifyResp User
				if err := json.Unmarshal(body, &verifyResp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if verifyResp.Username != "testuser" {
					t.Fatalf("expected username %q, got %q", "testuser", verifyResp.Username)
				}
				// Password hash should not be sent in response
				if verifyResp.PasswordHash != "" {
					t.Fatalf("expected PasswordHash to be empty in response")
				}
			} else {
				if !strings.Contains(bodyStr, tt.expectMessage) {
					t.Fatalf("expected error message %q in response, got %s", tt.expectMessage, bodyStr)
				}
			}
		})
	}
}

func TestGenerateAndVerifyToken(t *testing.T) {
	userID := 42
	secret := "very-secret"
	token := generateToken(userID, secret, 24)

	// Verify the token can be decoded
	decodedID := verifyToken(token, secret)
	if decodedID != userID {
		t.Fatalf("expected userID %d, got %d", userID, decodedID)
	}

	// Verify invalid token returns 0
	invalidToken := "not-a-token"
	if id := verifyToken(invalidToken, secret); id != -1 {
		t.Fatalf("expected -1 for invalid token, got %d", id)
	}
}
