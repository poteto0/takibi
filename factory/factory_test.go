package factory

import (
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

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

		ctx := takibi.NewContext[Bindings](nil, nil, &Bindings{Val: "initial"})
		next := func(c interfaces.IContext[Bindings]) error {
			return nil
		}

		err := mw(ctx, next)

		assert.Nil(t, err)
		assert.True(t, called)
		assert.Equal(t, "updated", ctx.Env().Val)
	})
}
