package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerCompanyRoutes(router *mux.Router) {
	router.HandleFunc("", s.handlecompanyCreate).Methods("POST")
	router.HandleFunc("/{id}/edit", s.handleCompanyUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleCompanyFind).Methods("GET")
}

func (s *Server) handlecompanyCreate(w http.ResponseWriter, r *http.Request) {
	var c dots.Company

	if ok := encodeJSON[dots.Company](w, r, &c); !ok {
		return
	}

	err := s.CompanyService.CreateCompany(r.Context(), &c)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.Company](w, r, http.StatusCreated, &c)
}

func (s *Server) handleCompanyUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.CompanyUpdate
	if err := json.NewDecoder(r.Body).Decode(&updata); err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "edit company: invalid json body"))
		return
	}

	u := dots.UserFromContext(r.Context())
	updata.Tid = &u.ID

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	c, err := s.CompanyService.UpdateCompany(r.Context(), id, &updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[dots.Company](w, r, http.StatusOK, c)
}

func (s *Server) handleCompanyFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.CompanyFilter
	ok := encodeJSON[dots.CompanyFilter](w, r, &filter)
	if !ok {
		return
	}

	ee, n, err := s.CompanyService.FindCompany(r.Context(), &filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondJSON[findCompanyResponse](w, r, http.StatusFound, &findCompanyResponse{Companys: ee, N: n})
}

type findCompanyResponse struct {
	Companys []*dots.Company `json:"entrY_types"`
	N        int             `json:"n"`
}
