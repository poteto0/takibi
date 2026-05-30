package takibi

import (
	"bytes"
	stdContext "context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestContext_Getter(t *testing.T) {
	type Bindings struct {
		Val string
	}

	t.Run("check context methods", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		bindings := &Bindings{Val: "test"}

		ctx := NewContext(w, req, bindings, nil)

		assert.Equal(t, bindings, ctx.Env())
		assert.Equal(t, req, ctx.Req().Raw())
		assert.Equal(t, w, ctx.Response())
	})
}

func TestContext_Reset(t *testing.T) {
	type Bindings struct {
		Val string
	}

	t.Run("check reset method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		bindings := &Bindings{Val: "test"}

		ctx := NewContext(w, req, bindings, nil)

		newReq := httptest.NewRequest(http.MethodPost, "/new", nil)
		newW := httptest.NewRecorder()

		ctx.Reset(newW, newReq)

		assert.Equal(t, bindings, ctx.Env())
		assert.Equal(t, newReq, ctx.Req().Raw())
		assert.Equal(t, newW, ctx.Response())
	})

	t.Run("pathParams cleared after reset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[Bindings](w, req, nil, nil)

		ctx.SetParam(map[string]string{"id": "42", "name": "foo"})
		assert.Equal(t, "42", ctx.ParamBy("id"))

		newReq := httptest.NewRequest(http.MethodGet, "/new", nil)
		ctx.Reset(httptest.NewRecorder(), newReq)

		assert.Empty(t, ctx.Param())
		assert.Equal(t, "", ctx.ParamBy("id"))
		assert.Equal(t, "", ctx.ParamBy("name"))
	})
}

func TestContext_Response(t *testing.T) {
	t.Run("check text method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil, nil)

		err := ctx.Text("hello world")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("check bytes method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil, nil)

		err := ctx.Bytes([]byte("hello world"))
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("check json method", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil, nil)

			err := ctx.Json(map[string]string{"msg": "hello"})
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"msg":"hello"}`, w.Body.String())
		})

		t.Run("fail if no content type", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil, nil)

			err := ctx.Json(map[string]string{"msg": "hello"})
			assert.Error(t, err)
			assert.Equal(t, "content-type must be application/json", err.Error())
		})

		t.Run("returns error when response is nil", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := NewContext[any](nil, req, nil, nil)

			err := ctx.Json(map[string]string{"msg": "hello"})
			assert.Error(t, err)
			assert.Equal(t, "response is nil", err.Error())
		})
	})

	t.Run("check status method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil, nil)

		err := ctx.Status(http.StatusCreated).Text("created")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "created", w.Body.String())
	})

	t.Run("check redirect method", func(t *testing.T) {
		tests := []struct {
			name    string
			path    string
			wantErr bool
		}{
			{"relative path", "/redirect", false},
			{"relative path with query", "/redirect?foo=bar", false},
			{"absolute URL rejected", "https://evil.example.com/steal", true},
			{"protocol-relative URL rejected", "//evil.example.com", true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()
				ctx := NewContext[any](w, req, nil, nil)

				err := ctx.Redirect(tt.path)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.Nil(t, err)
				assert.Equal(t, http.StatusFound, w.Code)
				assert.Equal(t, tt.path, w.Header().Get("Location"))
			})
		}

		t.Run("returns error when response is nil", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := NewContext[any](nil, req, nil, nil)

			err := ctx.Redirect("/safe")
			assert.Error(t, err)
			assert.Equal(t, "response is nil", err.Error())
		})
	})

	t.Run("check redirect external method", func(t *testing.T) {
		tests := []struct {
			name         string
			url          string
			allowedHosts []string
			wantErr      bool
		}{
			{
				name:         "allowed host redirects",
				url:          "https://api.example.com/callback",
				allowedHosts: []string{"api.example.com"},
			},
			{
				name:         "host not in allowlist",
				url:          "https://evil.example.com/steal",
				allowedHosts: []string{"api.example.com"},
				wantErr:      true,
			},
			{
				name:         "empty allowlist rejected",
				url:          "https://api.example.com/callback",
				allowedHosts: []string{},
				wantErr:      true,
			},
			{
				name:         "port is stripped when matching host",
				url:          "https://api.example.com:8080/path",
				allowedHosts: []string{"api.example.com"},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()
				ctx := NewContext[any](w, req, nil, nil)

				err := ctx.RedirectExternal(tt.url, tt.allowedHosts)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.Nil(t, err)
				assert.Equal(t, http.StatusFound, w.Code)
				assert.Equal(t, tt.url, w.Header().Get("Location"))
			})
		}

		t.Run("returns error when response is nil", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := NewContext[any](nil, req, nil, nil)

			err := ctx.RedirectExternal("https://api.example.com/cb", []string{"api.example.com"})
			assert.Error(t, err)
			assert.Equal(t, "response is nil", err.Error())
		})
	})

	t.Run("Stream data w/o write header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil, nil)

		for i := 0; i < 3; i++ {
			var buf bytes.Buffer
			buf.WriteString("data")

			err := ctx.Stream(buf.Bytes())
			assert.Nil(t, err)
			assert.Contains(t, "data", w.Body.String())
			w.Body.Reset()
		}
	})

	t.Run("render component", func(t *testing.T) {
		t.Run("render w/ component", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil, nil)

			component := templ.ComponentFunc(func(ctx stdContext.Context, w io.Writer) error {
				_, err := w.Write([]byte("Hello Takibi"))
				return err
			})
			config := &interfaces.RenderConfig{
				Component:   component,
				ContentType: "text/html",
			}

			err := ctx.Render(config)

			assert.Nil(t, err)
			assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
			assert.Equal(t, "Hello Takibi", w.Body.String())
		})

		t.Run("error w/ nil config", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil, nil)

			err := ctx.Render(nil)
			assert.Error(t, err)
			assert.Equal(t, "config is nil", err.Error())
		})

		t.Run("returns error when response is nil", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := NewContext[any](nil, req, nil, nil)

			component := templ.ComponentFunc(func(ctx stdContext.Context, w io.Writer) error {
				_, err := w.Write([]byte("Hello"))
				return err
			})
			err := ctx.Render(&interfaces.RenderConfig{Component: component})
			assert.Error(t, err)
			assert.Equal(t, "response is nil", err.Error())
		})
	})
}

func TestContext_Rq(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext[any](w, req, nil, nil)

	assert.Equal(t, req, ctx.Req().Raw())
}
