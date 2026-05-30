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
			expectedPathParam: map[string]string{},
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
			expectedPathParam: map[string]string{},
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

		// Assert — order is non-deterministic (map-based DFS), use set comparison
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
