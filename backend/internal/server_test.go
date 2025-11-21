package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(store Store) *Server {
	return &Server{store: store}
}

func TestShowsBoards(t *testing.T) {
	b := []Board{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	mock := &MockStore{
		ListBoardsFunc: func(ctx context.Context) ([]Board, error) {
			return b, nil
		},
	}
	s := newTestServer(mock)

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
