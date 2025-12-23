package interfaces

type IRouter[Bindings any] interface {
	/*
		find router by method &
		find node by path
	*/
	Find(method, path string) (INode[Bindings], []MiddlewareFunc[Bindings], map[string]string)

	Use(path string, middleware ...MiddlewareFunc[Bindings]) error

	/*
		Register GET method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Get(path string, handler HandlerFunc[Bindings]) error

	/*
		Register POST method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Post(path string, handler HandlerFunc[Bindings]) error

	/*
		Register PUT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Put(path string, handler HandlerFunc[Bindings]) error

	/*
		Register PATCH method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Patch(path string, handler HandlerFunc[Bindings]) error

	/*
		Register DELETE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Delete(path string, handler HandlerFunc[Bindings]) error

	/*
		Register HEAD method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Head(path string, handler HandlerFunc[Bindings]) error

	/*
		Register OPTIONS method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Options(path string, handler HandlerFunc[Bindings]) error

	/*
		Register TRACE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Trace(path string, handler HandlerFunc[Bindings]) error

	/*
		Register CONNECT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Connect(path string, handler HandlerFunc[Bindings]) error
}
