package internal

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestServer(store Store) *Server {
	tpls := map[string]*template.Template{
		"index": template.Must(template.New("index").Parse(`<html><body>{{range .Boards}}<div class="board">{{.Name}}</div>{{end}}</body></html>`)),
		"board": template.Must(template.New("board").Parse(`<html><body><h1>{{.Board.Name}}</h1>{{range .Columns}}{{range .Cards}}<div class="card">{{.Title}}</div>{{end}}{{end}}</body></html>`)),
	}
	return &Server{store: store, templates: tpls}
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

func TestCreateBoardHandler_PostCreatesAndRedirects(t *testing.T) {
	created := make(chan string, 1)
	mock := &MockStore{
		CreateBoardFunc: func(ctx context.Context, name string) (Board, error) {
			created <- name
			return Board{ID: 5, Name: name}, nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("name", "NewBoard")
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.createBoardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/" {
		t.Fatalf("expected Location '/', got %q", loc)
	}
	// verify store was called with the right name
	select {
	case name := <-created:
		if name != "NewBoard" {
			t.Fatalf("expected CreateBoard called with NewBoard, got %q", name)
		}
	default:
		t.Fatal("CreateBoard was not called")
	}
}

func TestBoardHandler_GetRendersBoardAndCards(t *testing.T) {
	mock := &MockStore{
		GetBoardByIDFunc: func(ctx context.Context, id int) (Board, error) {
			return Board{ID: id, Name: "MyBoard"}, nil
		},
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{
				{ID: 10, BoardID: boardID, Name: "To Do", Position: 1},
			}, nil
		},
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			return []Card{
				{ID: 100, ColumnID: columnID, Title: "Task 1", Description: "desc", Position: 1},
			}, nil
		},
	}
	s := newTestServer(mock)

	req := httptest.NewRequest(http.MethodGet, "/board/42", nil)
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	str := string(body)
	if !strings.Contains(str, "MyBoard") || !strings.Contains(str, "Task 1") {
		t.Fatalf("board page did not render expected content: %s", str)
	}
}

func TestHandleCardEdit_PostUpdatesAndRedirects(t *testing.T) {
	updated := make(chan Card, 1)
	mock := &MockStore{
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return Card{ID: id, ColumnID: 77, Title: "Old", Description: "old"}, nil
		},
		UpdateCardFunc: func(ctx context.Context, c Card) error {
			updated <- c
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return Column{ID: id, BoardID: 9, Name: "Col"}, nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("title", "NewTitle")
	form.Set("description", "NewDesc")
	req := httptest.NewRequest(http.MethodPost, "/card/123/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.cardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/board/9" {
		t.Fatalf("expected redirect to /board/9, got %q", loc)
	}
	select {
	case c := <-updated:
		if c.ID != 123 || c.Title != "NewTitle" || c.Description != "NewDesc" {
			t.Fatalf("UpdateCard called with unexpected card: %+v", c)
		}
	default:
		t.Fatal("UpdateCard was not called")
	}
}

// Additional tests for more handlers

func TestHandleCardMove_TargetColumnMovesAndRedirects(t *testing.T) {
	updated := make(chan Card, 1)
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			// pretend target column has one card so new pos becomes 2
			if columnID == 300 {
				return []Card{{ID: 201, ColumnID: 300, Title: "Existing", Position: 1}}, nil
			}
			return nil, nil
		},
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return Card{ID: id, ColumnID: 200, Title: "MoveMe", Position: 1}, nil
		},
		UpdateCardFunc: func(ctx context.Context, c Card) error {
			updated <- c
			return nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			// return column with board id for redirect
			if id == 300 {
				return Column{ID: 300, BoardID: 55, Name: "Target"}, nil
			}
			return Column{ID: id, BoardID: 55, Name: "Src"}, nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("target_column_id", "300")
	req := httptest.NewRequest(http.MethodPost, "/card/200/move", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.cardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/board/55" {
		t.Fatalf("expected redirect to /board/55, got %q", loc)
	}
	select {
	case c := <-updated:
		if c.ID != 200 || c.ColumnID != 300 || c.Position != 2 {
			t.Fatalf("UpdateCard called with unexpected card: %+v", c)
		}
	default:
		t.Fatal("UpdateCard was not called")
	}
}

func TestHandleCardMove_DirectionalMoveRight(t *testing.T) {
	updated := make(chan Card, 1)
	mock := &MockStore{
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return Card{ID: id, ColumnID: 10, Title: "Mover", Position: 1}, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			if id == 10 {
				return Column{ID: 10, BoardID: 99, Name: "A"}, nil
			}
			return Column{ID: id, BoardID: 99, Name: "B"}, nil
		},
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{
				{ID: 10, BoardID: boardID, Name: "A", Position: 1},
				{ID: 11, BoardID: boardID, Name: "B", Position: 2},
			}, nil
		},
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			// dest column 11 has one card already, so pos becomes 2
			if columnID == 11 {
				return []Card{{ID: 301, ColumnID: 11, Title: "X", Position: 1}}, nil
			}
			return nil, nil
		},
		UpdateCardFunc: func(ctx context.Context, c Card) error {
			updated <- c
			return nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("direction", "right")
	req := httptest.NewRequest(http.MethodPost, "/card/10/move", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.cardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/board/99" {
		t.Fatalf("expected redirect to /board/99, got %q", loc)
	}
	select {
	case c := <-updated:
		if c.ID != 10 || c.ColumnID != 11 || c.Position != 2 {
			t.Fatalf("UpdateCard called with unexpected card: %+v", c)
		}
	default:
		t.Fatal("UpdateCard was not called")
	}
}

func TestHandleCardDelete_DeletesAndRedirects(t *testing.T) {
	deleted := make(chan int, 1)
	mock := &MockStore{
		GetCardByIDFunc: func(ctx context.Context, id int) (Card, error) {
			return Card{ID: id, ColumnID: 44, Title: "DelMe", Position: 1}, nil
		},
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return Column{ID: id, BoardID: 66, Name: "Col"}, nil
		},
		DeleteCardFunc: func(ctx context.Context, id int) error {
			deleted <- id
			return nil
		},
	}
	s := newTestServer(mock)

	req := httptest.NewRequest(http.MethodPost, "/card/150/delete", nil)
	rr := httptest.NewRecorder()
	s.cardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/66" {
		t.Fatalf("expected redirect to /board/66, got %q", loc)
	}
	select {
	case id := <-deleted:
		if id != 150 {
			t.Fatalf("expected DeleteCard called with 150, got %d", id)
		}
	default:
		t.Fatal("DeleteCard was not called")
	}
}

func TestHandleBoardCreateColumn_AddsColumn(t *testing.T) {
	created := make(chan Column, 1)
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{
				{ID: 1, BoardID: boardID, Name: "Existing", Position: 1},
			}, nil
		},
		CreateColumnFunc: func(ctx context.Context, c Column) (Column, error) {
			created <- c
			// simulate created with ID
			c.ID = 77
			return c, nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("name", "NewCol")
	req := httptest.NewRequest(http.MethodPost, "/board/5/columns/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/5" {
		t.Fatalf("expected redirect to /board/5, got %q", loc)
	}
	select {
	case c := <-created:
		if c.BoardID != 5 || c.Name != "NewCol" || c.Position != 2 {
			t.Fatalf("CreateColumn called with unexpected column: %+v", c)
		}
	default:
		t.Fatal("CreateColumn was not called")
	}
}

func TestHandleColumnEdit_UpdatesAndRedirects(t *testing.T) {
	updated := make(chan Column, 1)
	mock := &MockStore{
		GetColumnByIDFunc: func(ctx context.Context, id int) (Column, error) {
			return Column{ID: id, BoardID: 7, Name: "OldName"}, nil
		},
		UpdateColumnFunc: func(ctx context.Context, c Column) error {
			updated <- c
			return nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("name", "Renamed")
	req := httptest.NewRequest(http.MethodPost, "/board/7/column/11/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/7" {
		t.Fatalf("expected redirect to /board/7, got %q", loc)
	}
	select {
	case c := <-updated:
		if c.ID != 11 || c.Name != "Renamed" {
			t.Fatalf("UpdateColumn called with unexpected column: %+v", c)
		}
	default:
		t.Fatal("UpdateColumn was not called")
	}
}

func TestHandleColumnMove_SwapsAndRedirects(t *testing.T) {
	updates := make(chan Column, 4)
	mock := &MockStore{
		GetColumnsByBoardIDFunc: func(ctx context.Context, boardID int) ([]Column, error) {
			return []Column{
				{ID: 20, BoardID: boardID, Name: "Left", Position: 1},
				{ID: 21, BoardID: boardID, Name: "Right", Position: 2},
			}, nil
		},
		UpdateColumnFunc: func(ctx context.Context, c Column) error {
			updates <- c
			return nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("direction", "right")
	req := httptest.NewRequest(http.MethodPost, "/board/7/column/20/move", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/7" {
		t.Fatalf("expected redirect to /board/7, got %q", loc)
	}
	// Expect two updates: the swapped columns
	count := 0
	for count < 2 {
		select {
		case c := <-updates:
			if c.ID != 20 && c.ID != 21 {
				t.Fatalf("unexpected UpdateColumn id: %d", c.ID)
			}
			count++
		default:
			t.Fatalf("expected UpdateColumn to be called twice, got %d", count)
		}
	}
}

func TestHandleColumnDelete_DeletesAndRedirects(t *testing.T) {
	deleted := make(chan int, 1)
	mock := &MockStore{
		DeleteColumnFunc: func(ctx context.Context, id int) error {
			deleted <- id
			return nil
		},
	}
	s := newTestServer(mock)

	req := httptest.NewRequest(http.MethodPost, "/board/7/column/11/delete", nil)
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/7" {
		t.Fatalf("expected redirect to /board/7, got %q", loc)
	}
	select {
	case id := <-deleted:
		if id != 11 {
			t.Fatalf("expected DeleteColumn called with 11, got %d", id)
		}
	default:
		t.Fatal("DeleteColumn was not called")
	}
}

func TestHandleAddCardToColumn_CreatesAndRedirects(t *testing.T) {
	created := make(chan Card, 1)
	mock := &MockStore{
		GetCardsByColumnIDFunc: func(ctx context.Context, columnID int) ([]Card, error) {
			// column empty => position 1
			return nil, nil
		},
		CreateCardFunc: func(ctx context.Context, c Card) (Card, error) {
			created <- c
			c.ID = 999
			return c, nil
		},
	}
	s := newTestServer(mock)

	form := url.Values{}
	form.Set("title", "NewTask")
	form.Set("description", "Details")
	req := httptest.NewRequest(http.MethodPost, "/board/7/column/11/cards/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/board/7" {
		t.Fatalf("expected redirect to /board/7, got %q", loc)
	}
	select {
	case c := <-created:
		if c.ColumnID != 11 || c.Title != "NewTask" || c.Position != 1 {
			t.Fatalf("CreateCard called with unexpected card: %+v", c)
		}
	default:
		t.Fatal("CreateCard was not called")
	}
}

func TestHandleBoardDelete_DeletesAndRedirects(t *testing.T) {
	deleted := make(chan int, 1)
	mock := &MockStore{
		DeleteBoardFunc: func(ctx context.Context, id int) error {
			deleted <- id
			return nil
		},
	}
	s := newTestServer(mock)

	req := httptest.NewRequest(http.MethodPost, "/board/9/delete", nil)
	rr := httptest.NewRecorder()
	s.boardHandler(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Fatalf("expected redirect to '/', got %q", loc)
	}
	select {
	case id := <-deleted:
		if id != 9 {
			t.Fatalf("expected DeleteBoard called with 9, got %d", id)
		}
	default:
		t.Fatal("DeleteBoard was not called")
	}
}
