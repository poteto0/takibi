package router

import (
	"net/http"
	"testing"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieRouter(t *testing.T) {
	tr := New[any]().(*trieRouter[any])

	assert.NotNil(t, tr)
	assert.NotNil(t, tr.trees)
	assert.Equal(t, len(SupportedHttpMethod), len(tr.trees))
}

func TestTrieRouter_add(t *testing.T) {
	t.Run("can add route into each trees", func(t *testing.T) {
		tr := New[any]().(*trieRouter[any])

		tests := []struct {
			name   string
			method string
			path   string
		}{
			{
				"Get",
				http.MethodGet,
				"/",
			},
			{
				"Get",
				http.MethodGet,
				"/users",
			},
			{
				"Post",
				http.MethodPost,
				"/users",
			},
			{
				"Put",
				http.MethodPut,
				"/users",
			},
			{
				"Patch",
				http.MethodPatch,
				"/users",
			},
			{
				"Delete",
				http.MethodDelete,
				"/users",
			},
			{
				"Head",
				http.MethodHead,
				"/users",
			},
			{
				"Options",
				http.MethodOptions,
				"/users",
			},
			{
				"Trace",
				http.MethodTrace,
				"/users",
			},
			{
				"Connect",
				http.MethodConnect,
				"/users",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := tr.add(test.method, test.path, func(ctx interfaces.IContext[any]) error {
					return nil
				})

				assert.Nil(t, err)
			})
		}
	})

	t.Run("error on duplicate added", func(t *testing.T) {
		tr := New[any]().(*trieRouter[any])

		err := tr.add(http.MethodGet, "/users", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.Nil(t, err)

		err = tr.add(http.MethodGet, "/users", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.ErrorIs(t, err, constants.ErrHandlerAlreadyExists)
	})

	t.Run("not supported method", func(t *testing.T) {
		tr := New[any]().(*trieRouter[any])

		err := tr.add("FOO", "/users", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.ErrorIs(t, err, constants.ErrNotSupportedMethod)
	})
}

func TestTrieRouter_Find(t *testing.T) {
	tr := New[any]().(*trieRouter[any])

	err := tr.Get("/users/:id/name", func(ctx interfaces.IContext[any]) error {
		return nil
	})
	assert.Nil(t, err)

	n, _, pathParam := tr.Find(http.MethodGet, "/users/123/name")
	assert.NotNil(t, n)
	assert.Equal(t, map[string]string{"id": "123"}, pathParam)
}

func TestTrieRouter_Use(t *testing.T) {
	tr := New[any]().(*trieRouter[any])
	mw := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }

	// Use on root
	err := tr.Use("/", mw)
	assert.Nil(t, err)

	// Check that middleware is added to all method trees
	for _, tree := range tr.trees {
		middlewares := tree.Middlewares()
		assert.Len(t, middlewares, 1)
	}

	// Use on specific path
	err = tr.Use("/api", mw)
	assert.Nil(t, err)
	// Note: Use("/api") creates the node /api in all trees if not exists.
	// Check one tree
	tree := tr.trees[http.MethodGet]
	node, middlewares, _ := tree.Find("/api")
	assert.NotNil(t, node)
	assert.Len(t, middlewares, 2) // root + api

	t.Run("if error on AddMiddleware, return error", func(t *testing.T) {
		err = tr.Use("api", mw)

		assert.Error(t, err)
	})
}

func TestTrieRouter_allMethods(t *testing.T) {
	tr := New[any]().(*trieRouter[any])

	t.Run("Get", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Get("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Post", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Post("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Put", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Put("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Patch", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Patch("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Delete", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Delete("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Head", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Head("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Options", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Options("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Trace", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Trace("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Connect", func(t *testing.T) {
		assert.Nil(
			t,
			tr.Connect("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})
}
