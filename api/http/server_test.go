package http_test

import (
	"testing"

	dotshttp "github.com/innermond/dots/http"
)

type Server struct {
	*dotshttp.Server
}

func MustOpenServer(t *testing.T) *Server {
	t.Helper()

	s := &Server{Server: dotshttp.NewServer()}

	return s
}
