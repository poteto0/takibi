package takibi

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		assert.Equal(t, req, ctx.Request())
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
		assert.Equal(t, newReq, ctx.Request())
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
}
