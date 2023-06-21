package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) registerUserRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleUserIndex).Methods("GET")
}

func (s *Server) handleUserIndex(w http.ResponseWriter, r *http.Request) {
	ses, err := s.getSession(r)
	if err != nil && err != http.ErrNoCookie {
		Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&ses)
}
