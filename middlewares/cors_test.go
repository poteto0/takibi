package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestCors(t *testing.T) {
	t.Run("default config sets origin", func(t *testing.T) {
		mw := middlewares.Cors[any]()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("specific origin allowed", func(t *testing.T) {
		mw := middlewares.Cors[any](middlewares.CorsConfig{
			AllowOrigins: []string{"http://example.com"},
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("preflight request", func(t *testing.T) {
		mw := middlewares.Cors[any]()

		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		var nextCalled bool
		err := mw(ctx, func(c interfaces.IContext[any]) error {
			nextCalled = true
			return nil
		})

		assert.Nil(t, err)
		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("allow credentials", func(t *testing.T) {
		mw := middlewares.Cors[any](middlewares.CorsConfig{
			AllowOrigins:     []string{"http://example.com"},
			AllowCredentials: true,
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("expose headers", func(t *testing.T) {
		mw := middlewares.Cors[any](middlewares.CorsConfig{
			AllowOrigins:  []string{"*"},
			ExposeHeaders: []string{"X-Custom-Header", "X-Another-Header"},
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, "X-Custom-Header, X-Another-Header", rec.Header().Get("Access-Control-Expose-Headers"))
	})
}
