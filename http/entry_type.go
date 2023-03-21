package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerEntryTypeRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleEntryTypeCreate).Methods("POST")
	router.HandleFunc("/{id}/edit", s.handleEntryTypeUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleEntryTypeFind).Methods("GET")
}

func (s *Server) handleEntryTypeCreate(w http.ResponseWriter, r *http.Request) {
	var et dots.EntryType

	if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "new entry type: invalid json body"))
		return
	}

	err := s.EntryTypeService.CreateEntryType(r.Context(), &et)
	if err != nil {
		Error(w, r, err)
		return
	}

	w.Header().Set("Content-TYpe", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(et); err != nil {
		LogError(r, err)
		return
	}
}

func (s *Server) handleEntryTypeUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.EntryTypeUpdate
	if err := json.NewDecoder(r.Body).Decode(&updata); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "edit entry type: invalid json body"))
		return
	}

	u := dots.UserFromContext(r.Context())
	updata.TID = &u.ID

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	et, err := s.EntryTypeService.UpdateEntryType(r.Context(), id, &updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.EntryType](w, r, http.StatusOK, et)
}

func (s *Server) handleEntryTypeFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.EntryTypeFilter
	ok := encodeJSON[dots.EntryTypeFilter](w, r, &filter)
	if !ok {
		return
	}

	ee, n, err := s.EntryTypeService.FindEntryType(r.Context(), &filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[findEntryTypeResponse](w, r, http.StatusFound, &findEntryTypeResponse{EntryTypes: ee, N: n})
}

type findEntryTypeResponse struct {
	EntryTypes []*dots.EntryType `json:"entrY_types"`
	N          int               `json:"n"`
}
