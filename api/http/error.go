package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/innermond/dots"
)

func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, message := dots.ErrorCode(err), dots.ErrorMessage(err)
	deverr := err
	logit := false
	if werr := errors.Unwrap(err); werr != nil {
		deverr = werr
		logit = true
	}
	if code == dots.EINTERNAL || logit {
		log.Printf("[http] error: %s %s %s", r.Method, r.URL.Path, deverr)
	}

	errorStatus := errorStatusFromCode(code)
	w.WriteHeader(errorStatus)

	switch r.Header.Get("Accept") {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&errorResponse{Error: message})
	default:
		w.Write([]byte(message))
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

var codes = map[string]int{
	dots.ECONFLICT:       http.StatusConflict,
	dots.EINVALID:        http.StatusBadRequest,
	dots.ENOTFOUND:       http.StatusNotFound,
	dots.ENOTIMPLEMENTED: http.StatusNotImplemented,
	dots.EUNAUTHORIZED:   http.StatusUnauthorized,
	dots.EINTERNAL:       http.StatusInternalServerError,
}

func errorStatusFromCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

func codeFromErrorStatus(status int) string {
	for k, v := range codes {
		if v == status {
			return k
		}
	}
	return dots.EINTERNAL
}
