package internal

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rhnvrm/simples3"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	addr        string
	store       Store
	secret      string
	expiry      int
	router      *chi.Mux
	finalized   bool
	s3Client    *simples3.S3
	s3ClientAlt *simples3.S3
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
		nil,
		nil,
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

func (s *Server) S3Client(client *simples3.S3) *Server {
	s.s3Client = client
	s.s3ClientAlt = S3ClientWithAlternateEndpoint(client)
	return s
}

// Docs registers the swagger route.
func (s *Server) Docs() *Server {
	if !s.finalized {
		s.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

// Auth registers the auth routes.
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

// Overview registers the overview routes.
func (s *Server) Overview() *Server {
	if !s.finalized {
		s.router.Get("/api/v2/boards", s.listBoards)
	} else {
		log.Fatal("Server is already finalized")
	}
	return s
}

// Board registers the board routes.
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

// Column registers the column routes.
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

// Card registers the card routes.
func (s *Server) Card() *Server {
	if !s.finalized {
		s.router.Route("/api/v2/card", func(r chi.Router) {
			r.Get("/{columnId}", s.getCards)
			r.Post("/", s.createCard)
			r.Put("/", s.updateCard)
			r.Delete("/", s.deleteCard)

			if s.s3Client != nil {
				log.Println("S3 client is configured, adding the attachment routes")
				r.Route("/{cardId}/attachment", func(r chi.Router) {
					r.Put("/prepare", s.prepareAttachment)
					r.Put("/{attachmentId}/finalize", s.finalizeAttachment)
					r.Delete("/{attachmentId}", s.deleteAttachment)
				})
			}
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
