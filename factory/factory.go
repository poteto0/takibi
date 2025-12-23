package factory

import (
	"github.com/poteto0/takibi/interfaces"
)

// CreateMiddleware creates a MiddlewareFunc from a function that takes a context and next handler, and returns a handler.
func CreateMiddleware[Bindings any](
	f func(interfaces.IContext[Bindings], interfaces.HandlerFunc[Bindings]) interfaces.HandlerFunc[Bindings],
) interfaces.MiddlewareFunc[Bindings] {
	return func(c interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) error {
		return f(c, next)(c)
	}
}
