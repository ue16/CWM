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
	columns := []Column{
		{ID: 1, Name: "Column 1", BoardID: 1},
		{ID: 2, Name: "Column 2", BoardID: 1},
	}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return columns, nil
		},
	}
	s := &Server{store: mock}

	req := httptest.NewRequest(http.MethodGet, "/column/1", nil)
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

func TestCreateColumn(t *testing.T) {
	column := Column{ID: 1, Name: "New Column", BoardID: 1, Position: 1}
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{}, nil
		},
		CreateColumnFunc: func(ctx context.Context, c Column) (Column, error) {
			return column, nil
		},
	}
	s := &Server{store: mock}

	body := `{"board_id": 1, "name": "New Column"}`
	req := httptest.NewRequest(http.MethodPost, "/column/", strings.NewReader(body))
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

func TestUpdateColumn(t *testing.T) {
	mock := &MockStore{
		UpdateColumnFunc: func(ctx context.Context, c Column) error {
			return nil
		},
	}
	s := &Server{store: mock}

	body := `{"id": 1, "name": "Updated Column"}`
	req := httptest.NewRequest(http.MethodPut, "/column/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.updateColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteColumn(t *testing.T) {
	mock := &MockStore{
		DeleteColumnFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}
	s := &Server{store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/column/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.deleteColumn(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}
