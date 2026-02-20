package interfaces

type HandlerFunc[Bindings any] = func(ctx IContext[Bindings]) error
type MiddlewareFunc[Bindings any] func(c IContext[Bindings], next HandlerFunc[Bindings]) error
type ErrorHandlerFunc[Bindings any] func(ctx IContext[Bindings], err error) error
type BlowErrorHandlerFunc[Bindings any] func(c IContext[Bindings], err error)
