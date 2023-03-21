package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerDeedRoutes(router *mux.Router) {
	router.HandleFunc("", s.handleDeedCreate).Methods("POST")
	router.HandleFunc("/{id}/edit", s.handleDeedUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleDeedFind).Methods("GET")
}

func (s *Server) handleDeedCreate(w http.ResponseWriter, r *http.Request) {
	var d dots.Deed

	if ok := inputJSON[dots.Deed](w, r, &d); !ok {
		return
	}

	err := s.DeedService.CreateDeed(r.Context(), &d)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.Deed](w, r, http.StatusCreated, &d)
}

func (s *Server) handleDeedUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.DeedUpdate
	if err := json.NewDecoder(r.Body).Decode(&updata); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "edit deed: invalid json body"))
		return
	}

	//u := dots.UserFromContext(r.Context())

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	d, err := s.DeedService.UpdateDeed(r.Context(), id, &updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.Deed](w, r, http.StatusOK, d)
}

func (s *Server) handleDeedFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.DeedFilter
	ok := inputJSON[dots.DeedFilter](w, r, &filter)
	if !ok {
		return
	}

	dd, n, err := s.DeedService.FindDeed(r.Context(), &filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[findDeedResponse](w, r, http.StatusFound, &findDeedResponse{Deeds: dd, N: n})
}

type findDeedResponse struct {
	Deeds []*dots.Deed `json:"deeds"`
	N     int          `json:"n"`
}
