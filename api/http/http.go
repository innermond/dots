package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/innermond/dots"
)

var ErrInputMissing = errors.New("missing input")

func LogError(r *http.Request, err error) {
	log.Printf("[http] %s %s %s", r.Method, r.URL.Path, err)
}

// inputjSON decodes JSON stream into a struct pointed by e param
func inputJSON[T any](w http.ResponseWriter, r *http.Request, e *T, prefix string) bool {
	// first check missing input
	if r.Body == http.NoBody {
		LogError(r, ErrInputMissing)
		msg := fmt.Sprintf("%s: empty input", prefix)
		Error(w, r, dots.Errorf(dots.EINVALID, msg))
		return false
	}

	// keep the body here
	buf := bytes.Buffer{}
	rb := io.NopCloser(io.TeeReader(r.Body, &buf))

	if err := json.NewDecoder(rb).Decode(e); err != nil {
		LogError(r, err)
		msg := fmt.Sprintf("%s: undecodable input", prefix)
		Error(w, r, dots.Errorf(dots.EINVALID, msg))
		return false
	}

	xx, err := unknownFieldsJSON(e, &buf)
	if err != nil {
		Error(w, r, err)
		return false
	}
	if len(xx) > 0 {
		msg := strings.Join(xx, ", ")
		// cut too long utf-8 string
		msgAsRunes := []rune(msg)
		cutAt := 200
		if len(msgAsRunes) > cutAt {
			msg = string(msgAsRunes[:cutAt])
		}
		Error(w, r, dots.Errorf(dots.ENOTFOUND, fmt.Sprintf("unknown input: %s", msg)))
		return false
	}

	return true
}

func unknownFieldsJSON(s interface{}, r io.Reader) ([]string, error) {
	var m map[string]interface{}
	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, err
	}

	v := reflect.ValueOf(s).Elem()
	t := v.Type()
	var unknownFields []string
	for k := range m {
		found := false
		for i := 0; i < t.NumField(); i++ {
			fieldName := t.Field(i).Name
			tagValue := t.Field(i).Tag.Get("json")
			if tagValue != "" {
				tagParts := strings.Split(tagValue, ",")
				if tagParts[0] == k {
					found = true
					break
				}
			}
			if fieldName == k {
				found = true
				break
			}
		}
		if !found {
			unknownFields = append(unknownFields, k)
		}
	}
	return unknownFields, nil
}

func outputJSON[T any](w http.ResponseWriter, r *http.Request, status int, response *T) {
	w.Header().Set("Content-TYpe", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		LogError(r, err)
	}
}
