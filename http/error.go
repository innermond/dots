package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/innermond/dots"
)

func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, message := dots.ErrorCode(err), dots.ErrorMessage(err)
	if code == dots.EINTERNAL {
		log.Printf("[http] error: %s %s %s", r.Method, r.URL.Path, err)
	}
	errorStatus := errorStatusFromCode(code)
	switch r.Header.Get("Accept") {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorStatus)
		json.NewEncoder(w).Encode(&errorResponse{Error: message})
	default:
		w.WriteHeader(errorStatus)
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

func codeFromErrorStatus(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return dots.EINTERNAL
}
