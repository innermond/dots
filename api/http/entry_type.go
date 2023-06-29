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
	router.HandleFunc("/{id}", s.handleEntryTypePatch).Methods("PATCH")
	router.HandleFunc("", s.handleEntryTypeFind).Methods("GET")
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

func (s *Server) handleEntryTypePatch(w http.ResponseWriter, r *http.Request) {
  if _, found := r.URL.Query()["del"]; found {
    s.handleEntryTypeDelete(w, r)
    return
  }

  s.handleEntryTypeUpdate(w, r)
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
  // can accept missing r.Body
	var filter dots.EntryTypeFilter

  // ensure we have a input body to be sent to json
	if r.Body != http.NoBody {
    buf := bytes.Buffer{}
    r.Body = io.NopCloser(io.TeeReader(r.Body, &buf))

    if ok := inputJSON(w, r, &filter, "find entry type"); !ok {
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
  }

	ee, n, err := s.EntryTypeService.FindEntryType(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &findEntryTypeResponse{EntryTypes: ee, N: n})
}

func (s *Server) handleEntryTypeDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

  filter := dots.EntryTypeDelete{}
  if r.Body != http.NoBody {
    ok := inputJSON(w, r, &filter, "delete entry type")
    if !ok {
      return
    }
  }

  if _, found := r.URL.Query()["resurect"]; found {
		filter.Resurect = true
	}
	n, err := s.EntryTypeService.DeleteEntryType(r.Context(), id, filter)
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
