package internal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	addr   string
	store  Store
	secret string
	expiry int
}

// Prepare creates and returns an http.Handler that renders simple Kanban pages using Go html/templates.
//
// Note: the function keeps the same parameters so existing call sites don't need to change,
// but it returns an http.Handler instead of starting the server. The caller can then
// start the server (e.g. http.ListenAndServe) or use the handler for testing.
func PrepareServer(addr, secret string, store Store) http.Handler {
	s := &Server{
		addr,
		store,
		secret,
		24,
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
		// Public auth routes
		r.Post("/auth/login", s.login)
		r.Post("/auth/register", s.register)
		r.Get("/auth/verify", s.verify)

		// Protected routes - boards
		r.Get("/boards", s.listBoards)
		r.Get("/board/{boardId}", s.getBoard)
		r.Post("/board/", s.createBoard)
		r.Put("/board/", s.updateBoard)
		r.Delete("/board/", s.deleteBoard)

		// Column actions
		r.Get("/column/{boardId}", s.getColumns)
		r.Post("/column/", s.createColumn)
		r.Put("/column/", s.updateColumn)
		r.Delete("/column/", s.deleteColumn)

		// Card actions
		r.Get("/card/{columnId}", s.getCards)
		r.Post("/card/", s.createCard)
		r.Put("/card/", s.updateCard)
		r.Delete("/card/", s.deleteCard)
	})

	// Return the configured handler for the caller to use (start server or test).
	return r
}
