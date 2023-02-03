package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) registerUserRoutes(router *mux.Router) {
	router.HandleFunc("/", s.handleUserIndex).Methods("GET")
}

func (s *Server) handleUserIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("user works!"))
}
