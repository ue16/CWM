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

func TestGetColumns(t *testing.T) {
	secret := "secret"
	columns := []Column{
		{ID: 1, Name: "Column 1", BoardID: 1, UserID: 1},
		{ID: 2, Name: "Column 2", BoardID: 1, UserID: 1},
	}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return columns, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareBoardOwnerFunc: func(ctx context.Context, boardID int, userID int) (bool, error) {
			return true, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/column/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("boardId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getColumns(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Response: %v", resp.StatusCode, resp)
	}
	body, _ := io.ReadAll(resp.Body)
	str := string(body)
	if expect, _ := json.Marshal(columns); strings.TrimSpace(string(expect)) != strings.TrimSpace(str) {
		t.Fatalf("response body mismatch: %s. Expected: %s", str, strings.TrimSpace(string(expect)))
	}
}

func TestGetColumnsUnauthorized(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{}, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareBoardOwnerFunc: func(ctx context.Context, boardID int, userID int) (bool, error) {
			return false, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/column/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("boardId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getColumns(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCreateColumn(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "New Column", BoardID: 1, Position: 1, UserID: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{}, nil
		},
		CreateColumnFunc: func(ctx context.Context, c Column) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareBoardOwnerFunc: func(ctx context.Context, boardID int, userID int) (bool, error) {
			return true, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"board_id": 1, "name": "New Column"}`
	req := httptest.NewRequest(http.MethodPost, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.createColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	str := string(respBody)
	if expect, _ := json.Marshal(column); strings.TrimSpace(string(expect)) != strings.TrimSpace(str) {
		t.Fatalf("response body mismatch: %s", str)
	}
}

func TestCreateColumnUnauthorized(t *testing.T) {
	secret := "secret"
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{}, nil
		},
		CreateColumnFunc: func(ctx context.Context, c Column) (Column, error) {
			return Column{}, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareBoardOwnerFunc: func(ctx context.Context, boardID int, userID int) (bool, error) {
			return false, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"board_id": 1, "name": "New Column"}`
	req := httptest.NewRequest(http.MethodPost, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.createColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateColumn(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Updated Column", BoardID: 1, Position: 1, UserID: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		UpdateColumnFunc: func(ctx context.Context, c Column) error {
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareColumnOwnerFunc: func(ctx context.Context, columnID int, userID int) (bool, error) {
			return true, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "name": "Updated Column"}`
	req := httptest.NewRequest(http.MethodPut, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateColumnUnauthorized(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Updated Column", BoardID: 1, Position: 1, UserID: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		UpdateColumnFunc: func(ctx context.Context, c Column) error {
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareColumnOwnerFunc: func(ctx context.Context, columnID int, userID int) (bool, error) {
			return false, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "name": "Updated Column"}`
	req := httptest.NewRequest(http.MethodPut, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeleteColumn(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Column to Delete", BoardID: 1, Position: 1, UserID: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		DeleteColumnFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareColumnOwnerFunc: func(ctx context.Context, columnID int, userID int) (bool, error) {
			return true, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestDeleteColumnUnauthorized(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Column to Delete", BoardID: 1, Position: 1, UserID: 2}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		DeleteColumnFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
		CompareColumnOwnerFunc: func(ctx context.Context, columnID int, userID int) (bool, error) {
			return false, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/column/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}
