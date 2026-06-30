package router

import (
	"reflect"
	"testing"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	n := NewNode[any]()

	assert.NotNil(t, n)
}

func TestNode_Handler(t *testing.T) {
	// Arrange
	n := NewNode[any]().(*node[any])
	handler := func(ctx interfaces.IContext[any]) error {
		return nil
	}
	n.handler = handler

	assert.Equal(t, reflect.TypeOf(n.Handler()), reflect.TypeOf(handler))
}

func TestNode_Add(t *testing.T) {
	t.Run("return error if handler already exists", func(t *testing.T) {
		n := NewNode[any]()

		err := n.Add("/foo", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.Nil(t, err)

		err = n.Add("/foo", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.ErrorIs(t, err, constants.ErrHandlerAlreadyExists)
	})

	t.Run("return error if handler already exists at /", func(t *testing.T) {
		n := NewNode[any]()

		err := n.Add("/", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.Nil(t, err)

		err = n.Add("/", func(ctx interfaces.IContext[any]) error {
			return nil
		})
		assert.ErrorIs(t, err, constants.ErrHandlerAlreadyExists)
	})
}

func TestNode_AddAndFind(t *testing.T) {
	// Arrange
	n := NewNode[any]()

	// Arrange & Assert
	err := n.Add("/", func(ctx interfaces.IContext[any]) error {
		return nil
	})
	assert.Nil(t, err)
	err = n.Add(
		"/foo/:bar",
		func(ctx interfaces.IContext[any]) error {
			return nil
		},
	)
	assert.Nil(t, err)
	err = n.Add(
		"/foo/:bar/baz",
		func(ctx interfaces.IContext[any]) error {
			return nil
		},
	)
	assert.Nil(t, err)

	tests := []struct {
		name              string
		path              string
		isNode            bool
		expectedPathParam map[string]string
	}{
		{
			name:              "root",
			path:              "/",
			isNode:            true,
			expectedPathParam: nil,
		},
		{
			name:   "foo",
			path:   "/foo/bar",
			isNode: true,
			expectedPathParam: map[string]string{
				"bar": "bar",
			},
		},
		{
			name:   "foo baz",
			path:   "/foo/bar/baz",
			isNode: true,
			expectedPathParam: map[string]string{
				"bar": "bar",
			},
		},
		{
			name:              "not found",
			path:              "/not-found",
			isNode:            false,
			expectedPathParam: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			n, _, pathParams := n.Find(test.path)

			// Assert
			assert.Equal(t, test.expectedPathParam, pathParams)
			assert.Equal(t, test.isNode, n != nil)
		})
	}
}

func TestNode_Find_LazyAllocation(t *testing.T) {
	t.Run("matched static route returns nil middlewares and pathParams", func(t *testing.T) {
		n := NewNode[any]()
		err := n.AddMiddleware("/", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil })
		assert.Nil(t, err)
		err = n.Add("/users", func(ctx interfaces.IContext[any]) error { return nil })
		assert.Nil(t, err)

		// The hot path: a matched route uses the pre-composed handler, so
		// Find allocates neither the middlewares slice nor the pathParams map.
		found, middlewares, pathParams := n.Find("/users")
		assert.NotNil(t, found)
		assert.NotNil(t, found.ComposedHandler())
		assert.Nil(t, middlewares)
		assert.Nil(t, pathParams)
	})

	t.Run("not found returns prefix middlewares for the 404 handler", func(t *testing.T) {
		n := NewNode[any]()
		err := n.AddMiddleware("/", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil })
		assert.Nil(t, err)
		err = n.AddMiddleware("/api", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil })
		assert.Nil(t, err)

		found, middlewares, _ := n.Find("/api/missing")
		assert.Nil(t, found)
		assert.Len(t, middlewares, 2) // root + api
	})
}

func TestNode_Middlewares(t *testing.T) {
	n := NewNode[any]()
	mw := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }

	err := n.AddMiddleware("/", mw)
	assert.Nil(t, err)

	assert.Len(t, n.Middlewares(), 1)
}

func TestNode_AddMiddleware(t *testing.T) {
	mw1 := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }
	mw2 := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }

	t.Run("add middleware to root", func(t *testing.T) {
		n := NewNode[any]()

		err := n.AddMiddleware("/", mw1)
		assert.Nil(t, err)
		assert.Len(t, n.Middlewares(), 1)
	})

	t.Run("add middleware to root on star pattern", func(t *testing.T) {
		n := NewNode[any]()

		err := n.AddMiddleware("*", mw1)
		assert.Nil(t, err)
		assert.Len(t, n.Middlewares(), 1)
	})

	t.Run("add middleware to child", func(t *testing.T) {
		n := NewNode[any]()

		err := n.AddMiddleware("/", mw1)
		assert.Nil(t, err)

		err = n.AddMiddleware("/api/users/:id/auth", mw2)
		assert.Nil(t, err)

		// Find the child node
		child, middlewares, _ := n.Find("/api/users/1/auth")
		assert.NotNil(t, child)
		// Middlewares collected should include root + child
		assert.Len(t, middlewares, 2)
	})

	t.Run("return error try to add middleware to child w/o \"/\" suffix", func(t *testing.T) {
		n := NewNode[any]()

		err := n.AddMiddleware("api", mw2)
		assert.ErrorIs(t, err, constants.ErrInvalidPath)
	})
}

func TestNode_ComposedHandler(t *testing.T) {
	t.Run("returns nil when no handler registered", func(t *testing.T) {
		n := NewNode[any]()
		assert.Nil(t, n.ComposedHandler())
	})

	t.Run("returns non-nil after Add", func(t *testing.T) {
		n := NewNode[any]()
		n.Add("/", func(ctx interfaces.IContext[any]) error { return nil })
		assert.NotNil(t, n.ComposedHandler())
	})

	t.Run("executes middleware in correct order before handler", func(t *testing.T) {
		n := NewNode[any]()
		order := []string{}

		n.AddMiddleware("/", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "mw1")
			return next(c)
		})
		n.Add("/", func(ctx interfaces.IContext[any]) error {
			order = append(order, "handler")
			return nil
		})

		n.ComposedHandler()(nil)
		assert.Equal(t, []string{"mw1", "handler"}, order)
	})

	t.Run("middleware added after Add still composes correctly", func(t *testing.T) {
		n := NewNode[any]()
		order := []string{}

		n.Add("/path", func(ctx interfaces.IContext[any]) error {
			order = append(order, "handler")
			return nil
		})
		n.AddMiddleware("/", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "mw")
			return next(c)
		})

		found, _, _ := n.Find("/path")
		found.ComposedHandler()(nil)
		assert.Equal(t, []string{"mw", "handler"}, order)
	})

	t.Run("ancestor and child middlewares both apply", func(t *testing.T) {
		n := NewNode[any]()
		order := []string{}

		n.AddMiddleware("/", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "root-mw")
			return next(c)
		})
		n.AddMiddleware("/api", func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
			order = append(order, "api-mw")
			return next(c)
		})
		n.Add("/api/users", func(ctx interfaces.IContext[any]) error {
			order = append(order, "handler")
			return nil
		})

		found, _, _ := n.Find("/api/users")
		found.ComposedHandler()(nil)
		assert.Equal(t, []string{"root-mw", "api-mw", "handler"}, order)
	})
}

func TestNode_Linearize(t *testing.T) {
	emptyHandler := func(ctx interfaces.IContext[any]) error { return nil }
	mw := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }

	t.Run("ancestor middleware is included in child NodeUnit", func(t *testing.T) {
		n := NewNode[any]()
		n.AddMiddleware("/", mw)
		n.Add("/hello", emptyHandler)

		units := n.Linearize()

		assert.Len(t, units, 1)
		assert.Equal(t, "/hello", units[0].Path)
		assert.Len(t, units[0].Middleware, 1, "root middleware should be inherited by /hello NodeUnit")
	})

	t.Run("multiple levels of ancestor middleware are all included", func(t *testing.T) {
		n := NewNode[any]()
		mw2 := func(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error { return nil }
		n.AddMiddleware("/", mw)
		n.AddMiddleware("/api", mw2)
		n.Add("/api/users", emptyHandler)

		units := n.Linearize()

		assert.Len(t, units, 1)
		assert.Len(t, units[0].Middleware, 2, "both root and /api middlewares should be in NodeUnit")
	})

	t.Run("Linearize all nodes", func(t *testing.T) {
		// Arrange
		n := NewNode[any]()

		n.Add("/", emptyHandler)
		n.Add("/users", emptyHandler)
		n.Add("/users/:id", emptyHandler)
		n.Add("/users/:id/profile", emptyHandler)
		n.Add("/posts", emptyHandler)

		// Act
		units := n.Linearize()

		// Assert
		expectedPaths := []string{
			"",
			"/users",
			"/users/:id",
			"/users/:id/profile",
			"/posts",
		}
		assert.Len(t, units, len(expectedPaths))
		actualPaths := make([]string, len(units))
		for i, unit := range units {
			actualPaths[i] = unit.Path
			assert.NotNil(t, unit.Handler)
		}
		assert.ElementsMatch(t, expectedPaths, actualPaths)
	})
}
