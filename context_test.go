package takibi

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

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

		ctx := NewContext(w, req, bindings)

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

		ctx := NewContext(w, req, bindings)

		newReq := httptest.NewRequest(http.MethodPost, "/new", nil)
		newW := httptest.NewRecorder()

		ctx.Reset(newW, newReq)

		assert.Equal(t, bindings, ctx.Env())
		assert.Equal(t, newReq, ctx.Req().Raw())
		assert.Equal(t, newW, ctx.Response())
	})
}

func TestContext_Response(t *testing.T) {
	t.Run("check text method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil)

		err := ctx.Text("hello world")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("check bytes method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil)

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
			ctx := NewContext[any](w, req, nil)

			err := ctx.Json(map[string]string{"msg": "hello"})
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"msg":"hello"}`, w.Body.String())
		})

		t.Run("fail if no content type", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil)

			err := ctx.Json(map[string]string{"msg": "hello"})
			assert.Error(t, err)
			assert.Equal(t, "content-type must be application/json", err.Error())
		})
	})

	t.Run("check status method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil)

		err := ctx.Status(http.StatusCreated).Text("created")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "created", w.Body.String())
	})

	t.Run("check redirect method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil)

		err := ctx.Redirect("/redirect")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/redirect", w.Header().Get("Location"))
	})

	t.Run("Stream data w/o write header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := NewContext[any](w, req, nil)

		for i := 0; i < 3; i++ {
			var buf bytes.Buffer
			buf.WriteString("data")

			err := ctx.Stream(buf.Bytes())
			assert.Nil(t, err)
			assert.Contains(t, "data", w.Body.String())
			w.Body.Reset()
		}
	})

	t.Run("render template", func(t *testing.T) {
		t.Run("render w/ template", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil)
			tmpl := template.Must(template.New("test").Parse("Hello {{.Name}}"))
			config := &interfaces.RenderConfig{
				Template:    tmpl,
				ContentType: "text/html",
			}

			err := ctx.Render(config, map[string]string{"Name": "Takibi"})

			assert.Nil(t, err)
			assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
			assert.Equal(t, "Hello Takibi", w.Body.String())
		})

		t.Run("render w/ rendererMap", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil)

			ctx.RegisterRenderer(
				map[string]*template.Template{
					"index": template.Must(template.New("test").Parse("Hello {{.Name}}")),
				},
			)

			config := &interfaces.RenderConfig{
				Key:         "index",
				ContentType: "text/html",
			}

			err := ctx.Render(config, map[string]string{"Name": "Takibi"})

			assert.Nil(t, err)
			assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
			assert.Equal(t, "Hello Takibi", w.Body.String())
		})

		t.Run("error if template not found", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil)
			config := &interfaces.RenderConfig{
				Key: "not-found",
			}

			err := ctx.Render(config, map[string]string{"Name": "Takibi"})
			assert.Error(t, err)
			assert.Equal(t, "template not found", err.Error())
		})

		t.Run("error w/ nil config", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			ctx := NewContext[any](w, req, nil)

			err := ctx.Render(nil, map[string]string{"Name": "Takibi"})
			assert.Error(t, err)
			assert.Equal(t, "config is nil", err.Error())
		})
	})
}

func TestContext_Rq(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext[any](w, req, nil)

	assert.Equal(t, req, ctx.Req().Raw())
}
