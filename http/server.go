package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	server *http.Server
	router *mux.Router
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	router := s.router.PathPrefix("/").Subrouter()
	router.HandleFunc("/", s.handleIndex).Methods("GET")
	return s
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "index!")
}

func (s *Server) ListenAndServe(domain string) error {
	return http.ListenAndServe(domain, s.router)
}
