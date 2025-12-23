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
