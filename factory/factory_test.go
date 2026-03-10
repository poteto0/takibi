package factory

import (
	"net/http"
	"net/http/httptest"
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

func TestParamBy(t *testing.T) {
	type Bindings struct{}

	t.Run("0 value", func(t *testing.T) {
		ctx := takibi.NewContext[Bindings](nil, nil, nil)
		ctx.SetParam(map[string]string{"id": ""})

		val, err := ParamBy[int](ctx, "id")
		assert.NoError(t, err)
		assert.Equal(t, 0, val) // zero value for int
	})

	t.Run("int", func(t *testing.T) {
		t.Run("parse", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"id": "123"})

			val, err := ParamBy[int](ctx, "id")
			assert.NoError(t, err)
			assert.Equal(t, 123, val)
		})

		t.Run("failed", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"id": "failed to parse"})

			_, err := ParamBy[int](ctx, "id")
			assert.Error(t, err)
		})
	})

	t.Run("int 64", func(t *testing.T) {
		t.Run("parse", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"id": "1234567890"})

			val, err := ParamBy[int64](ctx, "id")
			assert.NoError(t, err)
			assert.Equal(t, int64(1234567890), val)
		})

		t.Run("failed", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"id": "failed to parse"})

			_, err := ParamBy[int64](ctx, "id")
			assert.Error(t, err)
		})
	})

	t.Run("float 64", func(t *testing.T) {
		t.Run("parse", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"value": "3.14"})

			val, err := ParamBy[float64](ctx, "value")
			assert.NoError(t, err)
			assert.Equal(t, 3.14, val)
		})

		t.Run("failed", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"value": "failed to parse"})

			_, err := ParamBy[float64](ctx, "value")
			assert.Error(t, err)
		})
	})

	t.Run("bool", func(t *testing.T) {
		t.Run("parse", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"flag": "true"})

			val, err := ParamBy[bool](ctx, "flag")
			assert.NoError(t, err)
			assert.True(t, val)
		})

		t.Run("failed", func(t *testing.T) {
			ctx := takibi.NewContext[Bindings](nil, nil, nil)
			ctx.SetParam(map[string]string{"flag": "failed to parse"})

			_, err := ParamBy[bool](ctx, "flag")
			assert.Error(t, err)
		})
	})

	t.Run("string", func(t *testing.T) {
		ctx := takibi.NewContext[Bindings](nil, nil, nil)
		ctx.SetParam(map[string]string{"name": "takibi"})

		val, err := ParamBy[string](ctx, "name")
		assert.NoError(t, err)
		assert.Equal(t, "takibi", val)
	})

	t.Run("failed to parse custom type", func(t *testing.T) {
		type custom struct {
			Name string
		}

		ctx := takibi.NewContext[Bindings](nil, nil, nil)
		ctx.SetParam(map[string]string{"name": "failed to parse"})

		_, err := ParamBy[custom](ctx, "name")
		assert.Error(t, err)
	})
}

func TestQueryBy(t *testing.T) {
	type Bindings struct{}

	t.Run("int", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
		ctx := takibi.NewContext[Bindings](nil, req, nil)

		val, err := QueryBy[int](ctx, "page")
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
	})

	t.Run("bool", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?flag=true", nil)
		ctx := takibi.NewContext[Bindings](nil, req, nil)

		val, err := QueryBy[bool](ctx, "flag")
		assert.NoError(t, err)
		assert.True(t, val)
	})
}
