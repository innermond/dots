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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Server struct {
	server *http.Server
	router *mux.Router
	sc     *securecookie.SecureCookie

	ClientID     string
	ClientSecret string

	PingService dots.PingService
	UserService dots.UserService
	AuthService dots.AuthService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	// because it uses defer it must be called first
	// so its defer function will be the last in the stack, like a safety net
	s.router.Use(reportPanic)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	router := s.router.PathPrefix("/").Subrouter()
	router.HandleFunc("/", s.handleIndex).Methods("GET")
	router.HandleFunc("/panic", s.handleFakingPanic).Methods("GET")
	router.HandleFunc("/ping", s.handlePing).Methods("GET")

	router.HandleFunc("/user", s.handleUser).Methods("POST")

	s.registerAuthRoutes()

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
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}
}

func (s *Server) ListenAndServe(domain string) error {
	return http.ListenAndServe(domain, s.router)
}

func reportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				// do something with err
				w.Write([]byte(fmt.Errorf("panic: %w", err).Error()))
			}
		}()

		next.ServeHTTP(w, r)
	})
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
	err := s.setSession(w, Session{})
	if err != nil {
		Error(w, r, err)
		return
	}
	w.Write([]byte("index is working!"))
}

func (s *Server) handleFakingPanic(w http.ResponseWriter, r *http.Request) {
	panic("panic")
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	id := s.PingService.ById(r.Context())
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(id)
}

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) {
	isJson := strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json")
	if !isJson {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(""))
	}

	var u *dots.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dots.Errorf(dots.EINTERNAL, "http: decoding %v", err))
	}

	err = s.UserService.CreateUser(r.Context(), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dots.Errorf(dots.EINTERNAL, "user service: adding user %v", err))
	}
}
