package interfaces

import "net/http"

type HandlerFunc[Bindings any] = func(ctx IContext[Bindings]) error
type MiddlewareFunc[Bindings any] func(c IContext[Bindings], next HandlerFunc[Bindings]) error
type ErrorHandlerFunc[Bindings any] func(ctx IContext[Bindings], err error) error
type BlowErrorHandlerFunc[Bindings any] func(c IContext[Bindings], err error)

// EnvResolverFunc builds the Bindings exposed to ctx.Env() for a single request.
type EnvResolverFunc[Bindings any] func(r *http.Request) *Bindings
