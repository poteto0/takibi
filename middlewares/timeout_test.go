package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

type mockContext[Bindings any] struct {
	req  *http.Request
	res  http.ResponseWriter
	env  *Bindings
	done bool
}

func (m *mockContext[Bindings]) Env() *Bindings                { return m.env }
func (m *mockContext[Bindings]) Request() *http.Request        { return m.req }
func (m *mockContext[Bindings]) Response() http.ResponseWriter { return m.res }
func (m *mockContext[Bindings]) Reset(w http.ResponseWriter, r *http.Request) {
	m.res = w
	m.req = r
}
func (m *mockContext[Bindings]) Status(code int) interfaces.IContext[Bindings] { return m }
func (m *mockContext[Bindings]) Text(text string) error                        { return nil }
func (m *mockContext[Bindings]) Json(data any) error                           { return nil }
func (m *mockContext[Bindings]) Redirect(url string) error                     { return nil }

func TestTimeout(t *testing.T) {
	t.Run("context has deadline", func(t *testing.T) {
		timeout := time.Millisecond * 10
		mw := Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := &mockContext[any]{req: req, res: rec}

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
	})

	t.Run("context is cancelled after timeout", func(t *testing.T) {
		timeout := 10 * time.Millisecond
		mw := Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := &mockContext[any]{req: req, res: rec}

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			time.Sleep(20 * time.Millisecond)
			return nil
		})

		assert.ErrorIs(t, err, constants.ErrRequestTimeout)
	})

	t.Run("panic case, return internal error", func(t *testing.T) {
		timeout := 10 * time.Millisecond
		mw := Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := &mockContext[any]{req: req, res: rec}

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			panic("error")
		})

		assert.Error(t, err)
	})
}
