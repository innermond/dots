package http

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
)

func (s *Server) registerAuthRoutes() {
	s.router.HandleFunc("/oauth/github", s.handleOAuthGithub).Methods("GET")
}

func (s *Server) handleOAuthGithub(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		Error(w, r, err)
		return
	}

	state := make([]byte, 64)
	_, err = io.ReadFull(rand.Reader, state)
	if err != nil {
		Error(w, r, err)
		return
	}
	session.State = hex.EncodeToString(state)

	err = s.setSession(w, session)
	if err != nil {
		Error(w, r, err)
		return
	}

	authUrl := s.OAuth2Config().AuthCodeURL(session.State)
	http.Redirect(w, r, authUrl, http.StatusFound)
}
