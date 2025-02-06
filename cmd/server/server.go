package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/user"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

type Server struct {
	addr         string
	db           *sql.DB
	tokenService *auth.TokenService
}

func NewServer(addr string, db *sql.DB, tokenService *auth.TokenService) *Server {
	return &Server{
		addr:         addr,
		db:           db,
		tokenService: tokenService,
	}
}

func (s *Server) Start() error {
	router := chi.NewRouter()

	// strip trailing slashes at the end of the url
	// e.g. /users/1/ -> /users/1
	// this middleware should be applied to all routes
	// to ensure that the url is correctly formatted
	router.Use(chimiddleware.StripSlashes)

	router.Mount("/api/v1", s.v1Router()) // api version 1 subrouter

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.addr),
		Handler: router,
	}

	log.Printf("Server started at port %s\n", s.addr)

	return srv.ListenAndServe()
}

func (s *Server) v1Router() *chi.Mux {
	r := chi.NewRouter()

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// user feature
	userStore := user.NewStore(s.db)
	userService := user.NewService(userStore, s.tokenService)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(r)

	return r
}
