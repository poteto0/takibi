package interfaces

type INode[Bindings any] interface {
	Handler() HandlerFunc[Bindings]
	ComposedHandler() HandlerFunc[Bindings]
	Add(path string, handler HandlerFunc[Bindings]) (err error)
	// Find locates the node for path. On a matched route, callers use
	// n.ComposedHandler() (middlewares are already composed into it) and
	// middlewares is nil. middlewares is only populated on the not-found or
	// handler-less paths, where it holds the matched-prefix middlewares for a
	// fallback handler. pathParams is nil when the path has no parameters.
	Find(path string) (n INode[Bindings], middlewares []MiddlewareFunc[Bindings], pathParams map[string]string)

	Linearize() []NodeUnit[Bindings]

	Middlewares() []MiddlewareFunc[Bindings]
	AddMiddleware(path string, middleware ...MiddlewareFunc[Bindings]) error
}

type NodeUnit[Bindings any] struct {
	Path       string
	Handler    HandlerFunc[Bindings]
	Middleware []MiddlewareFunc[Bindings]
}
