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

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
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

func TestUnmarshall_StoresTypedBody(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/users",
		validator.Unmarshall(func(u User, c MyContext) (User, error) {
			if u.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return User{}, validator.ErrStop
			}
			return u, nil
		}),
		func(c MyContext) error {
			called = true
			u, ok := validator.Valid[User](c, validator.TargetJson)
			assert.True(t, ok)
			assert.Equal(t, "alice", u.Name)
			assert.Equal(t, 30, u.Age)
			return c.Status(http.StatusCreated).Json(u)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"alice","age":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUnmarshall_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/users",
		validator.Unmarshall(func(u User, c MyContext) (User, error) {
			if u.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return User{}, validator.ErrStop
			}
			return u, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"age":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "name required", w.Body.String())
}

func TestParam_StoresValidatedParam(t *testing.T) {
	app := takibi.New(&Bindings{})
	app.Get("/users/:id",
		validator.Param(func(p map[string]string, c MyContext) (string, error) {
			id := p["id"]
			if id == "" {
				c.Status(http.StatusUnprocessableEntity).Text("id required")
				return "", validator.ErrStop
			}
			return id, nil
		}),
		func(c MyContext) error {
			id, ok := validator.Valid[string](c, validator.TargetParam)
			assert.True(t, ok)
			return c.Text("user " + id)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user 42", w.Body.String())
}

func TestValid_ReturnsZeroWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := takibi.NewContext[Bindings](w, req, &Bindings{}, nil)

	data, ok := validator.Valid[PostForm](ctx, "form")
	assert.False(t, ok)
	assert.Equal(t, PostForm{}, data)
}
