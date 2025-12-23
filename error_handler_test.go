package takibi_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

type Bindings struct{}

func TestDefaultErrorHandler(t *testing.T) {
	app := takibi.New[Bindings](nil)
	app.Get("/error", func(ctx interfaces.IContext[Bindings]) error {
		return errors.New("something went wrong")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/error", nil)
	app.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "something went wrong", w.Body.String())
}

func TestCustomErrorHandler(t *testing.T) {
	app := takibi.New[Bindings](nil)

	customErr := errors.New("custom error")

	app.OnError(func(ctx interfaces.IContext[Bindings], err error) error {
		if errors.Is(err, customErr) {
			return ctx.Status(http.StatusBadRequest).Text("bad request: " + err.Error())
		}
		return ctx.Status(http.StatusInternalServerError).Text("internal error")
	})

	app.Get("/custom-error", func(ctx interfaces.IContext[Bindings]) error {
		return customErr
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/custom-error", nil)
	app.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "bad request: custom error", w.Body.String())
}
