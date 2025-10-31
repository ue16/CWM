package internal

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	db        *DB
	templates map[string]*template.Template
}

// Serve starts an HTTP server that renders simple Kanban pages using Go html/templates.
func Serve(addr string, db *DB) error {
	s := Server{
		db: db,
		templates: map[string]*template.Template{
			"index": template.Must(template.ParseFiles("templates/index.html.tmpl")),
			"board": template.Must(template.ParseFiles("templates/board.html.tmpl")),
		},
	}

	// Serve static assets (CSS, JS, images) from ./static/ at URL path /static/.
	// Place your extracted unified CSS at static/styles.css (or other assets under static/).
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Index handler (GET)
	http.HandleFunc("/", s.indexHandler)
	// Create board (POST)
	http.HandleFunc("/boards", s.createBoardHandler)

	// Board handler: /board/{id} (GET) and board-scoped POST actions (e.g. /board/{id}/column/...)
	http.HandleFunc("/board/", s.boardHandler)

	// Card-scoped POST actions (e.g. /card/{id}/edit)
	http.HandleFunc("/card/", s.cardHandler)

	// Start server.
	log.Printf("starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	boards, err := s.db.ListBoards(r.Context())
	if err != nil {
		http.Error(w, "failed to list boards", http.StatusInternalServerError)
		log.Printf("ListBoards: %v", err)
		return
	}
	if err := s.templates["index"].Execute(w, struct {
		Boards []Board
	}{Boards: boards}); err != nil {
		log.Printf("render index: %v", err)
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
	if _, err := s.db.CreateBoard(r.Context(), name); err != nil {
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
	board, err := s.db.GetBoardByID(ctx, id)
	if err != nil {
		http.Error(w, "board not found", http.StatusNotFound)
		log.Printf("GetBoardByID %d: %v", id, err)
		return
	}

	cols, err := s.db.GetColumnsByBoardID(ctx, id)
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
		cards, err := s.db.GetCardsByColumnID(ctx, c.ID)
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
