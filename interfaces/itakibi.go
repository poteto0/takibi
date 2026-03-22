package interfaces

import (
	stdContext "context"
	"html/template"
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

	// set renderer map for render method
	//
	// register phase
	//
	//  app.Renderer(map[string]*template.Template{
	//   "index": template.Must(template.New("index").Parse("Hello {{.Name}}")),
	//  })
	//
	// render phase
	//
	//  ctx.Render("index", "Takibi")
	//
	// then, rendered result is "Hello Takibi"
	Renderer(rendererMap map[string]*template.Template)

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

	/*
		Register All method Route at once

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	All(path string, handler HandlerFunc[Bindings]) error

	/*
		Register multiple method & path Route at once

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	On(methods, paths []string, handler HandlerFunc[Bindings]) error

	// Blow registers task
	Blow(tasks ...BlowTask[Bindings])

	// Camp simulates a request without starting the server
	Camp(method, path string, opts ...CampOption) ICampResponse
}
