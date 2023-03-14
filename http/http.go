package http

import (
	"encoding/json"
	"log"
	"net/http"
)

func LogError(r *http.Request, err error) {
	log.Printf("[http] %s %s %s", r.Method, r.URL.Path, err)
}

func respondJSON[T any](w http.ResponseWriter, r *http.Request, status int, response *T) {
	w.Header().Set("Content-TYpe", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		LogError(r, err)
	}
}
