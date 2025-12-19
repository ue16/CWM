package internal

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	addr      string
	store     Store
	secret    string
	expiry    int
	router    *chi.Mux
	finalized bool
}

// Prepare creates and returns a basic server without any functional routes.
// It needs to be set up with the correct routes through the
func BasicServer(addr, secret string, store Store) *Server {
	s := &Server{
		addr,
		store,
		secret,
		24,
		chi.NewRouter(),
		false,
	}

	// A good base middleware stack
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Cancels the request on timeout
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Get("/api/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("up"))
	})

	return s
}

func (s *Server) Docs() *Server {
	if !s.finalized {
		s.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

func (s *Server) Auth() *Server {
	if !s.finalized {
		s.router.Route("/api/v2/auth", func(r chi.Router) {
			r.Post("/login", s.login)
			r.Post("/register", s.register)
			r.Get("/verify", s.verify)
		})
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

func (s *Server) Overview() *Server {
	if !s.finalized {
		s.router.Get("/api/v2/boards", s.listBoards)
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

func (s *Server) Board() *Server {
	if !s.finalized {
		s.router.Route("/api/v2/board", func(r chi.Router) {
			r.Get("/{boardId}", s.getBoard)
			r.Post("/", s.createBoard)
			r.Put("/", s.updateBoard)
			r.Delete("/", s.deleteBoard)
		})
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

func (s *Server) Column() *Server {
	if !s.finalized {
		s.router.Route("/api/v2/column", func(r chi.Router) {
			r.Get("/{boardId}", s.getColumns)
			r.Post("/", s.createColumn)
			r.Put("/", s.updateColumn)
			r.Delete("/", s.deleteColumn)
		})
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

func (s *Server) Card() *Server {
	if !s.finalized {
		s.router.Route("/api/v2/card", func(r chi.Router) {
			r.Get("/{columnId}", s.getCards)
			r.Post("/", s.createCard)
			r.Put("/", s.updateCard)
			r.Delete("/", s.deleteCard)
		})
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

// Finalize sets the server as finalized and returns the configured handler for
// use in a ListenAndServe call.
// No further routes can be added after this.
func (s *Server) Finalize() http.Handler {
	s.finalized = true
	return s.router
}
