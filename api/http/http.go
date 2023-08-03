package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/innermond/dots"
)

var (
	ErrInputMissing   = errors.New("missing input")
	ErrInputWrongInto = errors.New("wrong target to decode into")
)

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
	rCopy := io.TeeReader(r, new(bytes.Buffer))
	for k := range m {
		found := false
		for i := 0; i < t.NumField(); i++ {
			fieldName := t.Field(i).Name
			fmt.Println(fieldName)
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

			// Check if the field is a nested struct
			if t.Field(i).Type.Kind() == reflect.Struct {
				nestedField := v.Field(i)
				nestedUnknownFields, err := unknownFieldsJSON(nestedField.Addr().Interface(), rCopy)
				if err != nil {
					return nil, err
				}
				unknownFields = append(unknownFields, nestedUnknownFields...)
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

// inputURL decodes query parameters to a struct pointed by s param
func inputURLQuery[T any](w http.ResponseWriter, r *http.Request, s *T, prefix string) bool {
	err := queryInto[T](r.URL.Query(), s)
	if err == nil {
		return true
	}

	LogError(r, err)
	msg := fmt.Sprintf("%s: undecodable input", prefix)
	Error(w, r, dots.Errorf(dots.EINVALID, msg))
	return false
}

// queryInto map query params to pointed struct
func queryInto[T any](qp url.Values, s *T) error {
	if len(qp) == 0 {
		return ErrInputMissing
	}

	t := reflect.TypeOf(s)
	e := t.Elem()

	if t.Kind() != reflect.Ptr && e.Kind() != reflect.Struct {
		return ErrInputWrongInto
	}

	v := reflect.ValueOf(s).Elem()
	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		fn := f.Tag.Get("json")
		if fn == "" {
			fn = f.Name
		}

		pv := qp.Get(fn)
		if pv == "" {
			continue
		}

		fv := v.Field(i)
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(pv)
		case reflect.Int:
			iv, err := strconv.Atoi(pv)
			if err != nil {
				return err
			}
			fv.SetInt(int64(iv))
		case reflect.Ptr:
			switch fv.Type().Elem().Kind() {
			case reflect.String:
				fv.Set(reflect.ValueOf(&pv))
			case reflect.Int:
				iv, err := strconv.Atoi(pv)
				if err != nil {
					return err
				}
				fv.Set(reflect.ValueOf(&iv))
			}
		}
	}
	return nil
}

type Filter interface {
	dots.CompanyFilter | dots.EntryTypeFilter | dots.EntryFilter | dots.DeedFilter
}

func input[T Filter](w http.ResponseWriter, r *http.Request, filterPtr *T, msg string) {
	if len(r.URL.Query()) > 0 {
		ok := inputURLQuery(w, r, filterPtr, msg)
		if !ok {
			return
		}
	}

	if r.Body != http.NoBody {
		ok := inputJSON(w, r, filterPtr, msg)
		if !ok {
			return
		}
	}
}

type affected struct {
	N int `json:"n"`
}

type data interface {
	dots.Company | dots.EntryType | dots.Entry | dots.Deed
}

type foundResponse[T data] struct {
	Data []*T `json:"data"`
	affected
}
