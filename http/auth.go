package http

import (
	"net/http"
)

func (s *Server) registerAuthRoutes() {
	s.router.HandleFunc("/oauth/github", s.handleOAuthGithub).Methods("GET")
}

func (s *Server) handleOAuthGithub(w http.ResponseWriter, r *http.Request) {

}
