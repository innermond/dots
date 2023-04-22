package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

func (s *Server) registerCompanyRoutes(router *mux.Router) {
	router.HandleFunc("", s.handlecompanyCreate).Methods("POST")
	router.HandleFunc("/{id}/edit", s.handleCompanyUpdate).Methods("PATCH")
	router.HandleFunc("", s.handleCompanyFind).Methods("GET")
	router.HandleFunc("", s.handleCompanyDelete).Methods("PATCH")
}

func (s *Server) handlecompanyCreate(w http.ResponseWriter, r *http.Request) {
	var c dots.Company

	if ok := inputJSON[dots.Company](w, r, &c, "create company"); !ok {
		return
	}

	err := s.CompanyService.CreateCompany(r.Context(), &c)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[dots.Company](w, r, http.StatusCreated, &c)
}

func (s *Server) handleCompanyUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		Error(w, r, dots.Errorf(dots.EINVALID, "invalid ID format"))
		return
	}

	var updata dots.CompanyUpdate
	if ok := inputJSON[dots.CompanyUpdate](w, r, &updata, "update company"); !ok {
		return
	}

	u := dots.UserFromContext(r.Context())
	updata.TID = &u.ID

	if err := updata.Valid(); err != nil {
		Error(w, r, err)
		return
	}

	c, err := s.CompanyService.UpdateCompany(r.Context(), id, updata)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[dots.Company](w, r, http.StatusOK, c)
}

func (s *Server) handleCompanyFind(w http.ResponseWriter, r *http.Request) {
	var filter dots.CompanyFilter
	ok := inputJSON[dots.CompanyFilter](w, r, &filter, "find company")
	if !ok {
		return
	}

	ee, n, err := s.CompanyService.FindCompany(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[findCompanyResponse](w, r, http.StatusFound, &findCompanyResponse{Companys: ee, N: n})
}

func (s *Server) handleCompanyDelete(w http.ResponseWriter, r *http.Request) {
	var filter dots.CompanyDelete
	ok := inputJSON[dots.CompanyDelete](w, r, &filter, "delete company")
	if !ok {
		return
	}

	if r.URL.Query().Get("resurect") != "" {
		filter.Resurect = true
	}
	n, err := s.CompanyService.DeleteCompany(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	outputJSON[deleteCompanyResponse](w, r, http.StatusFound, &deleteCompanyResponse{N: n})
}

type findCompanyResponse struct {
	Companys []*dots.Company `json:"companies"`
	N        int             `json:"n"`
}

type deleteCompanyResponse struct {
	N int `json:"n"`
}
