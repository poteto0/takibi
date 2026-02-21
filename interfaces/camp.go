package interfaces

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type CampOption func(*http.Request)

func Header(key, value string) CampOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

func Body(v any) CampOption {
	return func(r *http.Request) {
		if reader, ok := v.(io.Reader); ok {
			rc := io.NopCloser(reader)
			r.Body = rc
			return
		}

		b, _ := json.Marshal(v)
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	}
}
