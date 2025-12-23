package interfaces

type INode[Bindings any] interface {
	Handler() HandlerFunc[Bindings]
	Add(path string, handler HandlerFunc[Bindings]) (err error)
	Find(path string) (n INode[Bindings], middlewares []MiddlewareFunc[Bindings], pathParams map[string]string)

	Middlewares() []MiddlewareFunc[Bindings]
	AddMiddleware(path string, middleware ...MiddlewareFunc[Bindings]) error
}
