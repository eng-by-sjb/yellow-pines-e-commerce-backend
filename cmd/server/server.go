package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/user"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

type Server struct {
	addr string
	db   *sql.DB
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

func (s *Server) Start() error {
	router := chi.NewRouter()
	v1Router := chi.NewRouter() // api version 1 subrouter

	// strip trailing slashes at the end of the url
	// e.g. /users/1/ -> /users/1
	// this middleware should be applied to all routes
	// to ensure that the url is correctly formatted
	router.Use(chimiddleware.StripSlashes)

	// health check
	v1Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// user feature
	userStore := user.NewStore(s.db)
	userService := user.NewService(userStore)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(v1Router)

	// products feature

	router.Mount("/api/v1", v1Router)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.addr),
		Handler: router,
	}

	log.Printf("Server started at port %s\n", s.addr)

	return srv.ListenAndServe()
}
