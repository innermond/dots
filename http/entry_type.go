package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerEntryTypeRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleEntryTypeCreate).Methods("POST")
}

func (s *Server) handleEntryTypeCreate(w http.ResponseWriter, r *http.Request) {
	var et dots.EntryType

	if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid json body"))
		return
	}

	err := s.EntryTypeService.CreateEntryType(r.Context(), &et)
	if err != nil {
		Error(w, r, err)
		return
	}

	// response
	w.Header().Set("Content-TYpe", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(et); err != nil {
		LogError(r, err)
		return
	}
}
