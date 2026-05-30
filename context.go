package takibi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/thttp"
)

type context[Bindings any] struct {
	env        *Bindings
	request    interfaces.IRequest
	response   http.ResponseWriter
	statusCode int
	pathParams map[string]string
}

func NewContext[Bindings any](w http.ResponseWriter, r *http.Request, bindings *Bindings) interfaces.IContext[Bindings] {
	return &context[Bindings]{
		env:        bindings,
		request:    thttp.NewRequest(r),
		response:   w,
		statusCode: http.StatusOK,
		pathParams: make(map[string]string),
	}
}

func (c *context[Bindings]) Env() *Bindings {
	return c.env
}

func (c *context[Bindings]) Response() http.ResponseWriter {
	return c.response
}

func (c *context[Bindings]) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = thttp.NewRequest(r)
	c.response = w
	c.statusCode = http.StatusOK
	clear(c.pathParams)
}

func (c *context[Bindings]) Status(code int) interfaces.IContext[Bindings] {
	c.statusCode = code
	return c
}

func (c *context[Bindings]) Text(text string) error {
	if c.response == nil {
		return fmt.Errorf("response is nil")
	}
	c.response.Header().Set("Content-Type", "text/plain")
	c.response.WriteHeader(c.statusCode)
	_, err := c.response.Write([]byte(text))
	return err
}

func (c *context[Bindings]) Bytes(data []byte) error {
	if c.response == nil {
		return fmt.Errorf("response is nil")
	}
	c.response.Header().Set("Content-Type", "application/octet-stream")
	c.response.WriteHeader(c.statusCode)
	_, err := c.response.Write(data)
	return err
}

func (c *context[Bindings]) Stream(data []byte) error {
	if c.response == nil {
		return fmt.Errorf("response is nil")
	}

	_, err := c.response.Write(data)
	return err
}

func (c *context[Bindings]) Json(data any) error {
	if c.request.Raw() == nil || c.response == nil {
		return fmt.Errorf("request or response is nil")
	}
	contentType := c.request.Raw().Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("content-type must be application/json")
	}

	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(c.statusCode)
	return json.NewEncoder(c.response).Encode(data)
}

func (c *context[Bindings]) Redirect(path string) error {
	parsed, err := url.Parse(path)
	if err == nil && parsed.Host != "" {
		return fmt.Errorf("redirect: absolute URLs are not allowed, use RedirectExternal")
	}
	http.Redirect(c.response, c.request.Raw(), path, http.StatusFound)
	return nil
}

func (c *context[Bindings]) RedirectExternal(rawURL string, allowedHosts []string) error {
	if len(allowedHosts) == 0 {
		return fmt.Errorf("redirect: allowedHosts must not be empty")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("redirect: invalid URL: %w", err)
	}
	if slices.Contains(allowedHosts, parsed.Hostname()) {
		http.Redirect(c.response, c.request.Raw(), rawURL, http.StatusFound)
		return nil
	}
	return fmt.Errorf("redirect: host %q is not in the allowlist", parsed.Hostname())
}

func (c *context[Bindings]) Render(config *interfaces.RenderConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.ContentType != "" {
		c.response.Header().Set("Content-Type", config.ContentType)
	} else {
		c.response.Header().Set("Content-Type", "text/html")
	}

	if config.Component != nil {
		return config.Component.Render(c.request.Raw().Context(), c.response)
	}

	return fmt.Errorf("component not found")
}

func (c *context[Bindings]) Req() interfaces.IRequest {
	return c.request
}

func (c *context[Bindings]) Param() map[string]string {
	return c.pathParams
}

func (c *context[Bindings]) ParamBy(key string) string {
	return c.pathParams[key]
}

func (c *context[Bindings]) SetParam(params map[string]string) {
	clear(c.pathParams)
	for k, v := range params {
		c.pathParams[k] = v
	}
}
