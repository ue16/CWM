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
	b := []Board{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	mock := &MockStore{
		ListBoardsFunc: func(ctx context.Context) ([]Board, error) {
			return b, nil
		},
	}
	s := &Server{store: mock}

	req := httptest.NewRequest(http.MethodGet, "/boards", nil)
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

func TestGetBoard(t *testing.T) {
	board := Board{ID: 1, Name: "Alpha"}
	mock := &MockStore{
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{store: mock}

	req := httptest.NewRequest(http.MethodGet, "/board/1", nil)
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

func TestCreateBoard(t *testing.T) {
	board := Board{ID: 1, Name: "New Board"}
	mock := &MockStore{
		CreateBoardFunc: func(ctx context.Context, name string) (Board, error) {
			return board, nil
		},
	}
	s := &Server{store: mock}

	body := `{"name": "New Board"}`
	req := httptest.NewRequest(http.MethodPost, "/board/", strings.NewReader(body))
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

func TestUpdateBoard(t *testing.T) {
	mock := &MockStore{
		UpdateBoardFunc: func(ctx context.Context, b Board) error {
			return nil
		},
	}
	s := &Server{store: mock}

	body := `{"id": 1, "name": "Updated Board"}`
	req := httptest.NewRequest(http.MethodPut, "/board/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.updateBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteBoard(t *testing.T) {
	mock := &MockStore{
		DeleteBoardFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}
	s := &Server{store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/board/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.deleteBoard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}
