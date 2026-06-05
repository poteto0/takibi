package thttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/poteto0/takibi/constants"
)

type RequestOption struct {
	MaxBodyBytes int64
}

type Request struct {
	request      *http.Request
	maxBodyBytes int64
}

func NewRequest(r *http.Request, opt *RequestOption) *Request {
	req := &Request{request: r, maxBodyBytes: constants.DefaultMaxBodyBytes}
	if opt != nil && opt.MaxBodyBytes > 0 {
		req.maxBodyBytes = opt.MaxBodyBytes
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
	contentType := r.ContentType()
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	limited := http.MaxBytesReader(nil, r.request.Body, r.maxBodyBytes)
	if err := json.NewDecoder(limited).Decode(dest); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("request body is empty")
		}
		return err
	}
	return nil
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
