package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerEntryTypeRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleEntryTypeCreate).Methods("POST")
	router.HandleFunc("/{id}", s.handleEntryTypeUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleEntryTypeFind).Methods("GET")
	router.HandleFunc("", s.handleEntryTypeDelete).Methods("PATCH")
}

func (s *Server) handleEntryTypeCreate(w http.ResponseWriter, r *http.Request) {
	var et dots.EntryType

	if ok := inputJSON(w, r, &et, "create entry type"); !ok {
		return
	}

	err := s.EntryTypeService.CreateEntryType(r.Context(), &et)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusCreated, &et)
}

func (s *Server) handleEntryTypeUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.EntryTypeUpdate
	if ok := inputJSON(w, r, &updata, "edit entry type"); !ok {
		return
	}

	u := dots.UserFromContext(r.Context())
	updata.TID = &u.ID

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	et, err := s.EntryTypeService.UpdateEntryType(r.Context(), id, updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusOK, et)
}

func (s *Server) handleEntryTypeFind(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	r.Body = io.NopCloser(io.TeeReader(r.Body, &buf))

	var filter dots.EntryTypeFilter
	ok := inputJSON(w, r, &filter, "find entry type")
	if !ok {
		return
	}

	xx, err := unknownFieldsJSON(&filter, &buf)
	if err != nil {
		Error(w, r, err)
		return
	}
	if len(xx) > 0 {
		Error(w, r, dots.Errorf(dots.ENOTFOUND, fmt.Sprintf("unknown input: %s", strings.Join(xx, ", "))))
		return
	}

	ee, n, err := s.EntryTypeService.FindEntryType(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &findEntryTypeResponse{EntryTypes: ee, N: n})
}

func (s *Server) handleEntryTypeDelete(w http.ResponseWriter, r *http.Request) {
	var filter dots.EntryTypeDelete
	ok := inputJSON(w, r, &filter, "delete entry type")
	if !ok {
		return
	}

	if r.URL.Query().Get("resurect") != "" {
		filter.Resurect = true
	}
	n, err := s.EntryTypeService.DeleteEntryType(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &deleteEntryTypeResponse{N: n})
}

type findEntryTypeResponse struct {
	EntryTypes []*dots.EntryType `json:"entrY_types"`
	N          int               `json:"n"`
}

type deleteEntryTypeResponse struct {
	N int `json:"n"`
}
