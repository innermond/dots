package http

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/go-github/v32/github"
	"github.com/innermond/dots"
	"golang.org/x/oauth2"
)

func (s *Server) registerAuthRoutes() {
	s.router.HandleFunc("/oauth/github", s.handleOAuthGithub).Methods("GET")
	s.router.HandleFunc("/oauth/github/callback", s.handleOAuthGithubCallback).Methods("GET")
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

func (s *Server) handleOAuthGithubCallback(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		Error(w, r, err)
		return
	}

	code, state := r.FormValue("code"), r.FormValue("state")
	if state != session.State {
		Error(w, r, errors.New("oauth state mismatch"))
	}

	tok, err := s.OAuth2Config().Exchange(r.Context(), code)
	if err != nil {
		Error(w, r, fmt.Errorf("oauth exchange error: %s", err))
		return
	}

	client := github.NewClient(
		oauth2.NewClient(
			r.Context(),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: tok.AccessToken},
			),
		),
	)

	u, _, err := client.Users.Get(r.Context(), "")
	if err != nil {
		Error(w, r, fmt.Errorf("cannot fetch github user: %s", err))
		return
	} else if u.ID == nil {
		Error(w, r, errors.New("user ID not given by Github"))
		return
	}

	var name string
	if u.Name != nil {
		name = *u.Name
	} else if u.Login != nil {
		name = *u.Login
	}

	var email string
	if u.Email != nil {
		email = *u.Email
	}

	auth := &dots.Auth{
		Source:       dots.AuthSourceGithub,
		SourceID:     strconv.FormatInt(*u.ID, 10),
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		User:         &dots.User{Name: name, Email: email},
	}

	if !tok.Expiry.IsZero() {
		auth.Expiry = tok.Expiry
	}

	err = s.AuthService.CreateAuth(r.Context(), auth)
	if err != nil {
		Error(w, r, fmt.Errorf("http: cannot create auth: %s", err))
		return
	}

	redirectURL := session.RedirectURL

	session.UserID = auth.UserID
	session.RedirectURL = ""
	session.State = ""
	if err := s.setSession(w, session); err != nil {
		Error(w, r, fmt.Errorf("cannot set session cookie: %s", err))
		return
	}

	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)

}
