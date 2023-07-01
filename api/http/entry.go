package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerEntryRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleEntryCreate).Methods("POST")
	router.HandleFunc("/{id}", s.handleEntryPatch).Methods("PATCH")
	router.HandleFunc("", s.handleEntryFind).Methods("GET")
}

func (s *Server) handleEntryCreate(w http.ResponseWriter, r *http.Request) {
	var e dots.Entry

	if ok := inputJSON(w, r, &e, "create entry"); !ok {
		return
	}

	err := s.EntryService.CreateEntry(r.Context(), &e)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusCreated, &e)
}

func (s *Server) handleEntryPatch(w http.ResponseWriter, r *http.Request) {
	if _, found := r.URL.Query()["del"]; found {
		s.handleEntryDelete(w, r)
		return
	}

	s.handleEntryUpdate(w, r)
}

func (s *Server) handleEntryUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.EntryUpdate
	if err := json.NewDecoder(r.Body).Decode(&updata); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "edit entry: invalid json body"))
		return
	}

	//u := dots.UserFromContext(r.Context())

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	e, err := s.EntryService.UpdateEntry(r.Context(), id, updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusOK, e)
}

func (s *Server) handleEntryFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.EntryFilter
	if ok := inputJSON(w, r, &filter, "find entry"); !ok {
		return
	}

	ee, n, err := s.EntryService.FindEntry(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &findEntryResponse{Entries: ee, N: n})
}

func (s *Server) handleEntryDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	filter := dots.EntryDelete{}
  if _, found := r.URL.Query()["resurect"]; found {
		filter.Resurect = true
	}
	n, err := s.EntryService.DeleteEntry(r.Context(), id, filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &deleteEntryResponse{N: n})
}

type findEntryResponse struct {
	Entries []*dots.Entry `json:"entries"`
	N       int           `json:"n"`
}

type deleteEntryResponse struct {
	N int `json:"n"`
}
