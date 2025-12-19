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

func TestGetCards(t *testing.T) {
	secret := "secret"
	cards := []Card{
		{ID: 1, Title: "Card 1", ColumnID: 1},
		{ID: 2, Title: "Card 2", ColumnID: 1},
	}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			return cards, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/card/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("columnId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getCards(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Response body: %s", resp.StatusCode, func() string {
			body, _ := io.ReadAll(resp.Body)
			return string(body)
		}())
	}
	body, _ := io.ReadAll(resp.Body)
	str := strings.TrimSpace(string(body)) // Trim whitespace
	if expect, _ := json.Marshal(cards); strings.TrimSpace(string(expect)) != str {
		t.Fatalf("response body mismatch: %s", str)
	}
}

func TestGetCardsUnauthorized(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			return []Card{}, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	req := httptest.NewRequest(http.MethodGet, "/card/1", nil)
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("columnId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rr := httptest.NewRecorder()
	s.getCards(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCreateCard(t *testing.T) {
	secret := "secret"
	card := Card{ID: 1, Title: "New Card", ColumnID: 1, Position: 1}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			return []Card{}, nil
		},
		CreateCardFunc: func(ctx context.Context, c Card) (Card, error) {
			return card, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"column_id": 1, "title": "New Card", "description": "Description"}`
	req := httptest.NewRequest(http.MethodPost, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.createCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	str := strings.TrimSpace(string(respBody)) // Trim whitespace
	if expect, _ := json.Marshal(card); strings.TrimSpace(string(expect)) != str {
		t.Fatalf("response body mismatch: %s", str)
	}
}

func TestCreateCardUnauthorized(t *testing.T) {
	secret := "secret"
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			return []Card{}, nil
		},
		CreateCardFunc: func(ctx context.Context, c Card) (Card, error) {
			return Card{}, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"column_id": 1, "title": "New Card", "description": "Description"}`
	req := httptest.NewRequest(http.MethodPost, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.createCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateCard(t *testing.T) {
	secret := "secret"
	card := Card{ID: 1, Title: "Updated Card", ColumnID: 1, Description: "Updated Description", Position: 1}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		UpdateCardFunc: func(ctx context.Context, c Card) error {
			return nil
		},
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return card, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "title": "Updated Card", "description": "Updated Description"}`
	req := httptest.NewRequest(http.MethodPut, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateCardUnauthorized(t *testing.T) {
	secret := "secret"
	card := Card{ID: 1, Title: "Updated Card", ColumnID: 1, Description: "Updated Description", Position: 1}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		UpdateCardFunc: func(ctx context.Context, c Card) error {
			return nil
		},
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return card, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1, "title": "Updated Card", "description": "Updated Description"}`
	req := httptest.NewRequest(http.MethodPut, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.updateCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeleteCard(t *testing.T) {
	secret := "secret"
	card := Card{ID: 1, Title: "Card to Delete", ColumnID: 1, Position: 1}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 1}
	mock := &MockStore{
		DeleteCardFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return card, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestDeleteCardUnauthorized(t *testing.T) {
	secret := "secret"
	card := Card{ID: 1, Title: "Card to Delete", ColumnID: 1, Position: 1}
	column := Column{ID: 1, Name: "Test Column", BoardID: 1, Position: 1}
	board := Board{ID: 1, Name: "Test Board", UserID: 2}
	mock := &MockStore{
		DeleteCardFunc: func(ctx context.Context, id int) error {
			return nil
		},
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return card, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return column, nil
		},
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return board, nil
		},
	}
	s := &Server{secret: secret, store: mock}

	body := `{"id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/card/", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(1, secret, 24))
	rr := httptest.NewRecorder()
	s.deleteCard(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}
