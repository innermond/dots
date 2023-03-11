package http

import (
	"log"
	"net/http"
)

func LogError(r *http.Request, err error) {
	log.Printf("[http] %s %s %s", r.Method, r.URL.Path, err)
}
