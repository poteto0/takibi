package interfaces

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ICampResponse interface {
	StatusCode() int
	Raw() *http.Response
	Unmarshall(v any) error
	Json() (map[string]any, error)
}

type CampOption func(*http.Request)

func Header(key, value string) CampOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

func Body(v any) CampOption {
	return func(r *http.Request) {
		if reader, ok := v.(io.Reader); ok {
			r.Body = io.NopCloser(reader)
			r.ContentLength = -1 // size unknown
			return
		}

		b, _ := json.Marshal(v)
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		r.ContentLength = int64(len(b))
	}
}
