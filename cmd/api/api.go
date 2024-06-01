package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/go-chi/chi"
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

	// health check
	v1Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// user feature
	router.Mount("/api/v1", v1Router)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.addr),
		Handler: router,
	}

	fmt.Printf("Server started at %s\n", s.addr)

	return srv.ListenAndServe()
}
