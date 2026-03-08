package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	t.Run("context has deadline", func(t *testing.T) {
		timeout := time.Millisecond * 10
		mw := middlewares.Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			return nil
		})

		assert.Nil(t, err)
	})

	t.Run("context is cancelled after timeout", func(t *testing.T) {
		timeout := 10 * time.Millisecond
		mw := middlewares.Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			time.Sleep(20 * time.Millisecond)
			return nil
		})

		assert.ErrorIs(t, err, constants.ErrRequestTimeout)
	})

	t.Run("panic case, return internal error", func(t *testing.T) {
		timeout := 10 * time.Millisecond
		mw := middlewares.Timeout[any](timeout)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := takibi.NewContext[any](rec, req, nil)

		err := mw(ctx, func(c interfaces.IContext[any]) error {
			panic("error")
		})

		assert.Error(t, err)
	})
}
