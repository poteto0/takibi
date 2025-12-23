package takibi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/poteto0/takibi/interfaces"
)

type context[Bindings any] struct {
	env        *Bindings
	request    *http.Request
	response   http.ResponseWriter
	statusCode int
}

func NewContext[Bindings any](w http.ResponseWriter, r *http.Request, bindings *Bindings) interfaces.IContext[Bindings] {
	return &context[Bindings]{
		env:        bindings,
		request:    r,
		response:   w,
		statusCode: http.StatusOK,
	}
}

func (c *context[Bindings]) Env() *Bindings {
	return c.env
}

func (c *context[Bindings]) Request() *http.Request {
	return c.request
}

func (c *context[Bindings]) Response() http.ResponseWriter {
	return c.response
}

func (c *context[Bindings]) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response = w
	c.statusCode = http.StatusOK
}

func (c *context[Bindings]) Status(code int) interfaces.IContext[Bindings] {
	c.statusCode = code
	return c
}

func (c *context[Bindings]) Text(text string) error {
	c.response.Header().Set("Content-Type", "text/plain")
	c.response.WriteHeader(c.statusCode)
	_, err := c.response.Write([]byte(text))
	return err
}

func (c *context[Bindings]) Json(data any) error {
	contentType := c.request.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("content-type must be application/json")
	}

	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(c.statusCode)
	return json.NewEncoder(c.response).Encode(data)
}

func (c *context[Bindings]) Redirect(url string) error {
	http.Redirect(c.response, c.request, url, http.StatusFound)
	return nil
}
