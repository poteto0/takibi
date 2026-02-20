package factory

import (
	"context"
	"net/http"
	"testing"

	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

type mockContext[Bindings any] struct {
	env *Bindings
}

func (m *mockContext[Bindings]) Env() *Bindings                                { return m.env }
func (m *mockContext[Bindings]) Request() *http.Request                        { return nil }
func (m *mockContext[Bindings]) Response() http.ResponseWriter                 { return nil }
func (m *mockContext[Bindings]) Context() context.Context                      { return context.Background() }
func (m *mockContext[Bindings]) Reset(w http.ResponseWriter, r *http.Request)  {}
func (m *mockContext[Bindings]) Status(code int) interfaces.IContext[Bindings] { return m }
func (m *mockContext[Bindings]) Text(text string) error                        { return nil }
func (m *mockContext[Bindings]) Json(data any) error                           { return nil }
func (m *mockContext[Bindings]) Redirect(url string) error                     { return nil }

func TestCreateMiddleware(t *testing.T) {
	t.Run("creates middleware that executes logic", func(t *testing.T) {
		type Bindings struct {
			Val string
		}

		called := false

		mw := CreateMiddleware(func(c interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) interfaces.HandlerFunc[Bindings] {
			return func(c interfaces.IContext[Bindings]) error {
				called = true
				c.Env().Val = "updated"
				return next(c)
			}
		})

		ctx := &mockContext[Bindings]{env: &Bindings{Val: "initial"}}
		next := func(c interfaces.IContext[Bindings]) error {
			return nil
		}

		err := mw(ctx, next)

		assert.Nil(t, err)
		assert.True(t, called)
		assert.Equal(t, "updated", ctx.Env().Val)
	})
}
