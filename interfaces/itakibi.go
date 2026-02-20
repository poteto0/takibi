package interfaces

import (
	stdContext "context"
	"net/http"
)

type ITakibi[Bindings any] interface {
	// start up server
	Fire(addr string) error

	// stop server
	Finish(ctx stdContext.Context) error

	//
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	/* getter */
	Env() *Bindings

	OnError(handler ErrorHandlerFunc[Bindings])

	OnBlowError(handler BlowErrorHandlerFunc[Bindings])

	Use(path string, middleware ...MiddlewareFunc[Bindings]) error

	/* add node */
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

	// Blow registers a task
	Blow(task BlowTask[Bindings])
}
