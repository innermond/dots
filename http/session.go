package http

import (
	"net/http"
	"time"

	"github.com/segmentio/ksuid"
)

const SessionCookieName = "session"

type Session struct {
	UserID      ksuid.KSUID `json:"user_id"`
	RedirectURL string      `json:"redirect_url"`
	State       string      `json:"state"`
}

func (ses *Session) IsZero() bool {
	return ses.UserID == ksuid.Nil
}

func (s *Server) MarshalSession(session Session) (string, error) {
	return s.sc.Encode(SessionCookieName, session)
}

func (s *Server) UnmarshalSession(data string, session *Session) error {
	return s.sc.Decode(SessionCookieName, data, &session)
}

func (s *Server) setSession(w http.ResponseWriter, session Session) error {
	ss, err := s.MarshalSession(session)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    ss,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Secure:   false, // TODO change it when server will use ssl
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	// zero means deletion
	if session.IsZero() {
		cookie.MaxAge = 0
	}
	return nil
}

func (s *Server) getSession(r *http.Request) (Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = s.UnmarshalSession(cookie.Value, &session)
	if err != nil {
		return Session{}, err
	}
	return session, nil
}
