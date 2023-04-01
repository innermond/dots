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

	if ok := inputJSON[dots.EntryType](w, r, &et, "create entry type"); !ok {
		return
	}

	err := s.EntryTypeService.CreateEntryType(r.Context(), &et)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[dots.EntryType](w, r, http.StatusCreated, &et)
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

	outputJSON[dots.EntryType](w, r, http.StatusOK, et)
}

func (s *Server) handleEntryTypeFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.EntryTypeFilter
	ok := inputJSON[dots.EntryTypeFilter](w, r, &filter, "find entry type")
	if !ok {
		return
	}

	ee, n, err := s.EntryTypeService.FindEntryType(r.Context(), &filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[findEntryTypeResponse](w, r, http.StatusFound, &findEntryTypeResponse{EntryTypes: ee, N: n})
}

type findEntryTypeResponse struct {
	EntryTypes []*dots.EntryType `json:"entrY_types"`
	N          int               `json:"n"`
}
