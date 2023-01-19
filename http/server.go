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
	fmt.Fprint(w, []byte("index"))
}

func ListenAndServe(domain string) error {
	return http.ListenAndServe(domain+":8080", http.HandleFunc(s.router))
}
