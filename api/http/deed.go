package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerDeedRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleDeedCreate).Methods("POST")
	router.HandleFunc("/{id}/edit", s.handleDeedUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleDeedFind).Methods("GET")
	router.HandleFunc("", s.handleDeedDelete).Methods("PATCH")
}

func (s *Server) handleDeedCreate(w http.ResponseWriter, r *http.Request) {

	var d dots.Deed
	if ok := inputJSON(w, r, &d, "create deed"); !ok {
		return
	}

	err := s.DeedService.CreateDeed(r.Context(), &d)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusCreated, &d)
}

func (s *Server) handleDeedUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.DeedUpdate
	ok := inputJSON(w, r, &updata, "update deed")
	if !ok {
		return
	}

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	d, err := s.DeedService.UpdateDeed(r.Context(), id, updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusOK, d)
}

func (s *Server) handleDeedFind(w http.ResponseWriter, r *http.Request) {
  filter := dots.DeedFilter{}
	if r.Body != http.NoBody {
    ok := inputJSON(w, r, &filter, "find deed")
    if !ok {
      return
    }
  }

	dd, n, err := s.DeedService.FindDeed(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &findDeedResponse{Deeds: dd, N: n})
}

func (s *Server) handleDeedDelete(w http.ResponseWriter, r *http.Request) {
	var filter dots.DeedDelete
	ok := inputJSON(w, r, &filter, "delete deed")
	if !ok {
		return
	}

	if r.URL.Query().Get("resurect") != "" {
		filter.Resurect = true
	}
	n, err := s.DeedService.DeleteDeed(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON(w, r, http.StatusFound, &deleteDeedResponse{N: n})
}

type findDeedResponse struct {
	Deeds []*dots.Deed `json:"deeds"`
	N     int          `json:"n"`
}

type deleteDeedResponse struct {
	N int `json:"n"`
}
