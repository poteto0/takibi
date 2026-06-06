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

func TestSecureHeaders(t *testing.T) {
	run := func(mw interfaces.MiddlewareFunc[any]) (*httptest.ResponseRecorder, bool) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil, nil)
		var nextCalled bool
		_ = mw(ctx, func(c interfaces.IContext[any]) error {
			nextCalled = true
			return nil
		})
		return rec, nextCalled
	}

	t.Run("default config sets all security headers", func(t *testing.T) {
		rec, _ := run(middlewares.SecureHeaders[any]())
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	})

	t.Run("custom config overrides X-Frame-Options", func(t *testing.T) {
		rec, _ := run(middlewares.SecureHeaders[any](middlewares.SecureHeadersConfig{
			XFrameOptions: "SAMEORIGIN",
		}))
		assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
	})

	t.Run("empty X-Frame-Options skips the header", func(t *testing.T) {
		rec, _ := run(middlewares.SecureHeaders[any](middlewares.SecureHeadersConfig{
			XFrameOptions: "",
		}))
		assert.Empty(t, rec.Header().Get("X-Frame-Options"))
	})

	t.Run("calls next handler", func(t *testing.T) {
		_, nextCalled := run(middlewares.SecureHeaders[any]())
		assert.True(t, nextCalled)
	})
}
