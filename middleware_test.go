package takibi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	app := New[any](nil)

	// Global middleware
	app.Use("*", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
		c.Response().Header().Set("X-Global", "true")
		return next(c)
	})

	// Path middleware
	app.Use("/api/*", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
		c.Response().Header().Set("X-API", "true")
		return next(c)
	})

	app.Get("/", func(c interfaces.IContext[any]) error {
		return c.Text("root")
	})

	app.Get("/api/users", func(c interfaces.IContext[any]) error {
		return c.Text("users")
	})

	t.Run("Global middleware on root", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		assert.Equal(t, "true", rec.Header().Get("X-Global"))
		assert.Equal(t, "", rec.Header().Get("X-API"))
		assert.Equal(t, "root", rec.Body.String())
	})

	t.Run("Global and API middleware on /api/users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		assert.Equal(t, "true", rec.Header().Get("X-Global"))
		assert.Equal(t, "true", rec.Header().Get("X-API"))
		assert.Equal(t, "users", rec.Body.String())
	})

	t.Run("Middleware execution order", func(t *testing.T) {
		var order []string
		app2 := New[any](nil)

		app2.Use("*", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "global-pre")
			err := next(c)
			order = append(order, "global-post")
			return err
		})

		app2.Use("/api/*", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "api-pre")
			err := next(c)
			order = append(order, "api-post")
			return err
		})

		app2.Get("/api/test", func(c interfaces.IContext[any]) error {
			order = append(order, "handler")
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()
		app2.ServeHTTP(rec, req)

		expected := []string{"global-pre", "api-pre", "handler", "api-post", "global-post"}
		assert.Equal(t, expected, order)
	})
}
