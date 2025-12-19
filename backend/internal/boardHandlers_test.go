package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestListBoards(t *testing.T) {
	secret := "secret"
	b := []Board{
		{ID: 1, Name: "Alpha", UserID: 1},
		{ID: 2, Name: "Beta", UserID: 1},
	}
	mock := &MockStore{
		ListBoardsFunc: func(ctx context.Context, userID int) ([]Board, error) {
			return b, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/boards", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.listBoards(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	str := string(body)
	if expect, _ := json.Marshal(b); string(expect) != str {
		t.Fatalf("response body missing board names: %s", str)
	}
}

func TestListBoardsMissingHeader(t *testing.T) {
	secret := "secret"
	mock := &MockStore{
		ListBoardsFunc: func(ctx context.Context, userID int) ([]Board, error) {
			return nil, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/boards", nil)
	rr := httptest.NewRecorder()
	s.listBoards(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestGetBoard(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Alpha", UserID: 1}
	mock := &MockStore{
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/board/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("boardId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Response body: %s", resp.StatusCode, rr.Body.String())
	}
	body, _ := io.ReadAll(resp.Body)
	str := strings.TrimSpace(string(body))
	if expect, _ := json.Marshal(board); strings.TrimSpace(string(expect)) != str {
		t.Fatalf("response body mismatch: %s. Expected: %s", str, strings.TrimSpace(string(expect)))
	}
}

func TestGetBoardUnauthorized(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Alpha", UserID: 2}
	mock := &MockStore{
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/board/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("boardId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCreateBoard(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "New Board", UserID: 1}
	mock := &MockStore{
		CreateBoardFunc: func(ctx context.Context, name string, userID int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"name": "New Board"}`
	req := httptest.NewRequest(http.MethodPost, "/board/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.createBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	str := strings.TrimSpace(string(respBody))
	if expect, _ := json.Marshal(board); strings.TrimSpace(string(expect)) != str {
		t.Fatalf("response body mismatch: %s != %s", str, string(expect))
	}
}

func TestCreateBoardMissingHeader(t *testing.T) {
	secret := "secret"
	mock := &MockStore{
		CreateBoardFunc: func(ctx context.Context, name string, userID int) (Board, error) {
			return Board{}, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"name": "New Board"}`
	req := httptest.NewRequest(http.MethodPost, "/board/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.createBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestUpdateBoard(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Updated Board", UserID: 1}
	mock := &MockStore{
		UpdateBoardFunc: func(ctx context.Context, b Board) error {
			return nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "name": "Updated Board", "user_id": 1}`
	req := httptest.NewRequest(http.MethodPut, "/board/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateBoardUnauthorized(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Updated Board", UserID: 2}
	mock := &MockStore{
		UpdateBoardFunc: func(ctx context.Context, b Board) error {
			return nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "name": "Updated Board", "user_id": 2}`
	req := httptest.NewRequest(http.MethodPut, "/board/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeleteBoard(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Board to Delete", UserID: 1}
	mock := &MockStore{
		DeleteBoardFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/board/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestDeleteBoardUnauthorized(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Board to Delete", UserID: 2}
	mock := &MockStore{
		DeleteBoardFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/board/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}
