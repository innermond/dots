package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerEntryRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleEntryCreate).Methods("POST")
}

func (s *Server) handleEntryCreate(w http.ResponseWriter, r *http.Request) {
	var e dots.Entry

	if ok := encodeJSON[dots.Entry](w, r, &e); !ok {
		return
	}

	err := s.EntryService.CreateEntry(r.Context(), &e)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.Entry](w, r, http.StatusCreated, &e)
}
