package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/innermond/dots"
)

func LogError(r *http.Request, err error) {
	log.Printf("[http] %s %s %s", r.Method, r.URL.Path, err)
}

// inputjSON decodes JSON stream into a struct pointed by e param
func inputJSON[T any](w http.ResponseWriter, r *http.Request, e *T, prefix string) bool {
	if err := json.NewDecoder(r.Body).Decode(e); err != nil {
		msg := fmt.Sprintf("%s: the supplied input cannot be decoded", prefix)
		Error(w, r, dots.Errorf(dots.EINVALID, msg))
		return false
	}

	return true
}

func outputJSON[T any](w http.ResponseWriter, r *http.Request, status int, response *T) {
	w.Header().Set("Content-TYpe", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		LogError(r, err)
	}
}
