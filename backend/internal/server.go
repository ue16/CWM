package internal

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	addr      string
	store     Store
	templates map[string]*template.Template
}

// Prepare creates and returns an http.Handler that renders simple Kanban pages using Go html/templates.
//
// Note: the function keeps the same parameters so existing call sites don't need to change,
// but it returns an http.Handler instead of starting the server. The caller can then
// start the server (e.g. http.ListenAndServe) or use the handler for testing.
func PrepareServer(addr string, store Store) http.Handler {
	s := &Server{
		addr,
		store,
		nil,
	}
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Cancels the request on timeout
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("up"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/boards", s.listBoards)
	})

	// Return the configured handler for the caller to use (start server or test).
	return r
}

// @Summary	List all boards
// @Description Lists all existing boards.
// @Router		/boards [get]
// @Success	200	{array}	internal.Board
func (s *Server) listBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	boards, err := s.store.ListBoards(r.Context())
	if err != nil {
		http.Error(w, "{\"err\": \"failed to list boards\"", http.StatusInternalServerError)
		return
	}

	if boards, err := json.Marshal(boards); err == nil {
		w.Write(boards)
	} else {
		http.Error(w, "{\"err\": \"failed to marshal boards\"}", http.StatusInternalServerError)
		return
	}
}

func (s *Server) createBoardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if _, err := s.store.CreateBoard(r.Context(), name); err != nil {
		http.Error(w, "failed to create board", http.StatusInternalServerError)
		log.Printf("CreateBoard: %v", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// cardHandler dispatches card-scoped POST actions to dedicated handler functions.
func (s *Server) cardHandler(w http.ResponseWriter, r *http.Request) {
	// Handles POST actions on /card/{id}/...
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/card/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.SplitN(path, "/", 2)
	idStr := parts[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "edit":
		s.handleCardEdit(w, r, id)
		return
	case "move":
		s.handleCardMove(w, r, id)
		return
	case "delete":
		s.handleCardDelete(w, r, id)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

// boardHandler accepts GET for rendering the board page and POST for board/column scoped actions,
// delegating actions to dedicated handler functions.
func (s *Server) boardHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/board/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.SplitN(path, "/", 2)
	idStr := parts[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// POST actions under /board/{id}/...
	if r.Method == http.MethodPost {
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}

		// Board-scoped actions
		switch {
		case action == "delete":
			s.handleBoardDelete(w, r, id)
			return
		case action == "columns/add":
			s.handleBoardCreateColumn(w, r, id)
			return
		case strings.HasPrefix(action, "column/"):
			rest := strings.TrimPrefix(action, "column/")
			parts2 := strings.SplitN(rest, "/", 2)
			colIDStr := parts2[0]
			colID, err := strconv.Atoi(colIDStr)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			act := ""
			if len(parts2) > 1 {
				act = parts2[1]
			}
			// Dispatch column-scoped actions
			switch act {
			case "edit":
				s.handleColumnEdit(w, r, id, colID)
				return
			case "move":
				s.handleColumnMove(w, r, id, colID)
				return
			case "delete":
				s.handleColumnDelete(w, r, id, colID)
				return
			case "cards/add":
				s.handleAddCardToColumn(w, r, id, colID)
				return
			default:
				http.NotFound(w, r)
				return
			}
		default:
			http.NotFound(w, r)
			return
		}
	}

	// Default: GET rendering of the board page (existing behavior)
	ctx := r.Context()
	board, err := s.store.GetBoardByID(ctx, id)
	if err != nil {
		http.Error(w, "board not found", http.StatusNotFound)
		log.Printf("GetBoardByID %d: %v", id, err)
		return
	}

	cols, err := s.store.GetColumnsByBoardID(ctx, id)
	if err != nil {
		http.Error(w, "failed to load columns", http.StatusInternalServerError)
		log.Printf("GetColumnsByBoardID %d: %v", id, err)
		return
	}

	// Build view model: columns with nested cards.
	type ColWithCards struct {
		Column Column
		Cards  []Card
	}
	var colsWith []ColWithCards
	for _, c := range cols {
		cards, err := s.store.GetCardsByColumnID(ctx, c.ID)
		if err != nil {
			http.Error(w, "failed to load cards", http.StatusInternalServerError)
			log.Printf("GetCardsByColumnID %d: %v", c.ID, err)
			return
		}
		colsWith = append(colsWith, ColWithCards{Column: c, Cards: cards})
	}

	view := struct {
		Board   Board
		Columns []ColWithCards
	}{Board: board, Columns: colsWith}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates["board"].Execute(w, view); err != nil {
		log.Printf("render board %d: %v", id, err)
	}
}
