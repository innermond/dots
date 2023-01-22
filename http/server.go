package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/innermond/dots"
)

type Server struct {
	server *http.Server
	router *mux.Router

	pingService dots.PingService
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

	return s
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("index works!"))
}

func (s *Server) handleFakingPanic(w http.ResponseWriter, r *http.Request) {
	panic("panic")
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
				w.Write([]byte("panic: error"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
