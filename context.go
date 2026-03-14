package takibi

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/thttp"
)

type context[Bindings any] struct {
	env        *Bindings
	request    interfaces.IRequest
	response   http.ResponseWriter
	statusCode int
	pathParams map[string]string

	rendererMap map[string]*template.Template
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
	c.pathParams = make(map[string]string)
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

func (c *context[Bindings]) Steam(data []byte) error {
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

func (c *context[Bindings]) Redirect(url string) error {
	http.Redirect(c.response, c.request.Raw(), url, http.StatusFound)
	return nil
}

func (c *context[Bindings]) Render(config *interfaces.RenderConfig, data any) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.ContentType != "" {
		c.response.Header().Set("Content-Type", config.ContentType)
	} else {
		c.response.Header().Set("Content-Type", "text/html")
	}

	if config.IsTemplate() {
		return config.Template.Execute(c.response, data)
	}

	if tmpl, exists := c.rendererMap[config.Key]; exists {
		return tmpl.Execute(c.response, data)
	}
	return fmt.Errorf("template not found")
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

func (c *context[Bindings]) RegisterRenderer(rendererMap map[string]*template.Template) {
	c.rendererMap = rendererMap
}
