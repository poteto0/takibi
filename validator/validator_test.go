package validator_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/validator"
	"github.com/stretchr/testify/assert"
)

type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

type PostForm struct {
	Body string
}

func TestForm_StoresValidatedData(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			return PostForm{Body: v.Get("body")}, nil
		}),
		func(c MyContext) error {
			called = true
			data, ok := validator.Valid[PostForm](c, "form")
			assert.True(t, ok)
			assert.Equal(t, "hello", data.Body)
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	form := url.Values{"body": {"hello"}}
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestForm_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			if v.Get("body") == "" {
				c.Status(http.StatusUnprocessableEntity).Text("Invalid!")
				return PostForm{}, validator.ErrStop
			}
			return PostForm{Body: v.Get("body")}, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "Invalid!", w.Body.String())
}

func TestValid_ReturnsZeroWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := takibi.NewContext[Bindings](w, req, &Bindings{}, nil)

	data, ok := validator.Valid[PostForm](ctx, "form")
	assert.False(t, ok)
	assert.Equal(t, PostForm{}, data)
}
