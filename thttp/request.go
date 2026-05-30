package thttp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const DefaultMaxBodyBytes int64 = 10 << 20 // 10 MiB

type RequestOption func(*Request)

func WithMaxBodyBytes(n int64) RequestOption {
	return func(r *Request) { r.maxBodyBytes = n }
}

type Request struct {
	request      *http.Request
	maxBodyBytes int64
}

func NewRequest(r *http.Request, opts ...RequestOption) *Request {
	req := &Request{request: r, maxBodyBytes: DefaultMaxBodyBytes}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

func (r *Request) Raw() *http.Request {
	return r.request
}

func (r *Request) Header() http.Header {
	return r.request.Header
}

func (r *Request) HeaderBy(key string) string {
	return r.request.Header.Get(key)
}

func (r *Request) ContentType() string {
	return r.request.Header.Get("Content-Type")
}

func (r *Request) Json() (map[string]any, error) {
	var data map[string]any
	err := r.Unmarshall(&data)
	return data, err
}

func (r *Request) Unmarshall(dest any) error {
	if r.request.ContentLength == 0 {
		return fmt.Errorf("request body is empty")
	}

	if r.ContentType() != "application/json" {
		return fmt.Errorf("unsupported content type: %s", r.ContentType())
	}

	limited := http.MaxBytesReader(nil, r.request.Body, r.maxBodyBytes)
	return json.NewDecoder(limited).Decode(dest)
}

func (r *Request) Queries() map[string][]string {
	return r.request.URL.Query()
}

func (r *Request) QueriesBy(key string) []string {
	return r.request.URL.Query()[key]
}

func (r *Request) Query() map[string]string {
	raw := r.request.URL.Query()
	query := make(map[string]string, len(raw))
	for k, v := range raw {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}
	return query
}

func (r *Request) QueryBy(key string) string {
	values := r.request.URL.Query()[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
