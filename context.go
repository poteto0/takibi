package takibi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/thttp"
)

type contextOption struct {
	maxBodyBytes int64
}

func toContextOption(opt *interfaces.TakibiOption) contextOption {
	if opt == nil || opt.MaxBodyBytes == 0 {
		return contextOption{maxBodyBytes: constants.DefaultMaxBodyBytes}
	}
	return contextOption{maxBodyBytes: opt.MaxBodyBytes}
}

type context[Bindings any] struct {
	env           *Bindings
	request       interfaces.IRequest
	response      http.ResponseWriter
	statusCode    int
	pathParams    map[string]string
	maxBodyBytes  int64
	validatedData map[string]any
}

func NewContext[Bindings any](w http.ResponseWriter, r *http.Request, bindings *Bindings, opt *interfaces.TakibiOption) interfaces.IContext[Bindings] {
	co := toContextOption(opt)
	return &context[Bindings]{
		env:          bindings,
		request:      thttp.NewRequest(r, &thttp.RequestOption{MaxBodyBytes: co.maxBodyBytes}),
		response:     w,
		statusCode:   http.StatusOK,
		pathParams:   make(map[string]string),
		maxBodyBytes: co.maxBodyBytes,
	}
}

func (c *context[Bindings]) Env() *Bindings {
	return c.env
}

func (c *context[Bindings]) Response() http.ResponseWriter {
	return c.response
}

func (c *context[Bindings]) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = thttp.NewRequest(r, &thttp.RequestOption{MaxBodyBytes: c.maxBodyBytes})
	c.response = w
	c.statusCode = http.StatusOK
	clear(c.pathParams)
	c.validatedData = nil
}

func (c *context[Bindings]) SetValidated(target string, value any) {
	if c.validatedData == nil {
		c.validatedData = make(map[string]any)
	}
	c.validatedData[target] = value
}

func (c *context[Bindings]) Validated(target string) (any, bool) {
	v, ok := c.validatedData[target]
	return v, ok
}

func (c *context[Bindings]) Status(code int) interfaces.IContext[Bindings] {
	c.statusCode = code
	return c
}

func (c *context[Bindings]) checkResponse() error {
	if c.response == nil {
		return fmt.Errorf("response is nil")
	}
	return nil
}

func (c *context[Bindings]) Text(text string) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
	c.response.Header().Set("Content-Type", "text/plain")
	c.response.WriteHeader(c.statusCode)
	_, err := c.response.Write([]byte(text))
	return err
}

func (c *context[Bindings]) Bytes(data []byte) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
	c.response.Header().Set("Content-Type", "application/octet-stream")
	c.response.WriteHeader(c.statusCode)
	_, err := c.response.Write(data)
	return err
}

func (c *context[Bindings]) Stream(data []byte) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
	_, err := c.response.Write(data)
	return err
}

func (c *context[Bindings]) Json(data any) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(c.statusCode)
	return json.NewEncoder(c.response).Encode(data)
}

func (c *context[Bindings]) Redirect(path string) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
	parsed, err := url.Parse(path)
	if err == nil && parsed.Host != "" {
		return fmt.Errorf("redirect: absolute URLs are not allowed, use RedirectExternal")
	}
	http.Redirect(c.response, c.request.Raw(), path, http.StatusFound)
	return nil
}

func (c *context[Bindings]) RedirectExternal(rawURL string, allowedHosts []string) error {
	if err := c.checkResponse(); err != nil {
		return err
	}
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
	if err := c.checkResponse(); err != nil {
		return err
	}
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
	c.pathParams = params
}
