package http

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	server *http.Server
	router *mux.Router
	sc     *securecookie.SecureCookie

	ClientID     string
	ClientSecret string

	UserService  dots.UserService
	AuthService  dots.AuthService
	TokenService dots.TokenService

	EntryTypeService dots.EntryTypeService
	EntryService     dots.EntryService
	DrainService     dots.DrainService
	CompanyService   dots.CompanyService
	DeedService      dots.DeedService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter().PathPrefix("/v1").Subrouter(),
	}

	// because it uses defer it must be called first
	// so its defer function will be the last in the stack, like a safety net
	s.router.Use(reportPanic)
	s.router.Use(s.allowRequestsFromApp)
	s.server.Handler = http.HandlerFunc(s.serveHTTP)
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)
	s.router.Use(s.authenticate)

	s.router.Methods("OPTIONS")

	s.router.HandleFunc("/", s.handleIndex).Methods("GET")

	{
		router := s.router.PathPrefix("/").Subrouter()
		router.Use(s.noAuthenticate)
		s.registerAuthRoutes(router)
	}

	{
		router := s.router.PathPrefix("/me").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerUserRoutes(router)
	}

	{
		router := s.router.PathPrefix("/entry-types").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerEntryTypeRoutes(router)
	}

	{
		router := s.router.PathPrefix("/entries").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerEntryRoutes(router)
	}

	{
		router := s.router.PathPrefix("/drains").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerDrainRoutes(router)
	}

	{
		router := s.router.PathPrefix("/companies").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerCompanyRoutes(router)
	}

	{
		router := s.router.PathPrefix("/deeds").Subrouter()
		router.Use(s.yesAuthenticate)
		s.registerDeedRoutes(router)
	}

	return s
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		RedirectURL: "http://localhost:8080/oauth/google/callback",
		Endpoint:    google.Endpoint,
	}
}

func (s *Server) ListenAndServe(domain string) error {
	return http.ListenAndServe(domain, s.router)
}

func (s *Server) OpenSecureCookie() error {
	f, err := os.OpenFile(".securecookie", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bb := [][]byte{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		b, err := hex.DecodeString(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		bb = append(bb, b)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if len(bb) != 2 {
		log.Fatal("securecookie file: unexpected length")
	}

	s.sc = securecookie.New(bb[0], bb[1])
	s.sc.SetSerializer(securecookie.JSONEncoder{})
	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	ses, err := s.getSession(r)
	if err != nil && err != http.ErrNoCookie {
		Error(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&ses)
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{}
	resp["error"] = fmt.Sprintf("[server] '%s' not found", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(&resp)
}

func reportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				// do something with err
				w.Write([]byte(fmt.Errorf("panic: %v", err).Error()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func extractBearer(h string) string {
	hh := strings.SplitN(h, " ", 2)
	if len(hh) != 2 || strings.ToLower(hh[0]) != "bearer" {
		return ""
	}

	return hh[1]
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// try bearer
		// TODO check using return in middleware if breaks the chain
		tokenMaybe := extractBearer(r.Header.Get("Authorization"))
		if tokenMaybe != "" {
			payload, err := s.TokenService.Read(r.Context(), tokenMaybe)
			if err != nil {
				Error(w, r, err)
				return
			}
			if payload.UID != ksuid.Nil {
				u, err := s.UserService.FindUserByID(r.Context(), payload.UID)
				if err == nil {
					r = r.WithContext(dots.NewContextWithUser(r.Context(), u))
				} else {
					log.Printf("cannot find payload user %s: %s", payload.UID, err)
				}
			}
			next.ServeHTTP(w, r)
			return
		}
		ses, _ := s.getSession(r)
		if ses.UserID != ksuid.Nil {
			u, err := s.UserService.FindUserByID(r.Context(), ses.UserID)
			if err == nil {
				r = r.WithContext(dots.NewContextWithUser(r.Context(), u))
			} else {
				log.Printf("cannot find session user %s: %s", ses.UserID, err)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) yesAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := dots.UserFromContext(r.Context())
		if !u.ID.IsNil() {
			next.ServeHTTP(w, r)
			return
		}

		redirectURL := r.URL
		redirectURL.Scheme, redirectURL.Host = "", ""
		ses, err := s.getSession(r)
		if err != nil {
			log.Printf("cannot get session: %s", err)
		}
		ses.RedirectURL = redirectURL.String()
		err = s.setSession(w, ses)
		if err != nil {
			log.Printf("cannot set session: %s", err)
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func (s *Server) noAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := dots.UserFromContext(r.Context())
		isLogout := r.URL.Path == "/logout"
		if u.ID != ksuid.Nil && !isLogout {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) allowRequestsFromApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://www.dots.volt.com")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
