package interfaces

import (
	stdContext "context"
	"net/http"
)

type ITakibi[Bindings any] interface {
	// start up server
	//	- on native go:
	//		- start up net/http server
	// 	- on wasm:
	//		- use syumai worker
	Fire(addr string) error

	// stop server
	//	- ! it is not supported for wasm
	Finish(ctx stdContext.Context) error

	//
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	/* getter */
	Env() *Bindings

	OnError(handler ErrorHandlerFunc[Bindings])

	// OnBlowError sets the handler invoked when a Blow task returns an error.
	//	- on wasm: applies to "schedule" tasks dispatched by Cron Triggers.
	OnBlowError(handler BlowErrorHandlerFunc[Bindings])

	Use(path string, middleware ...MiddlewareFunc[Bindings]) error

	// Route registers sub app
	//
	// EX:
	//  api := takibi.New[any](nil)
	//  api.Get("/users", func(ctx interfaces.IContext[any]) error {
	//   return ctx.Text("users")
	//  })
	//
	//  app.Route("/api", api)
	//
	// then, GET /api/users will return "users"
	//
	//	- the sub app's error handler set by OnError is inherited: errors from
	//	  its handlers go to it, and an error it returns falls through to the
	//	  parent's error handler.
	//	- ! the sub app's Bindings are discarded: ctx.Env() always returns
	//	  the parent's Bindings, so pass every binding to the parent app.
	//	- ! the sub app's Blow tasks and OnBlowError handler are not merged.
	Route(basePath string, app ITakibi[Bindings]) error

	/* add node */
	/*
		Register GET method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Get(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register POST method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Post(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register PUT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Put(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register PATCH method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Patch(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register DELETE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Delete(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register HEAD method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Head(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register OPTIONS method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Options(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register TRACE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Trace(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register CONNECT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	Connect(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register All method Route at once

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	All(path string, handlers ...HandlerFunc[Bindings]) error

	/*
		Register multiple method & path Route at once

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	On(methods, paths []string, handlers ...HandlerFunc[Bindings]) error

	// Blow registers task
	//	- on native go: "trigger" (start/stop) and "schedule" tasks via robfig/cron.
	//	- on wasm: only "schedule" tasks, dispatched by Cloudflare Cron Triggers.
	//	  The firing schedule is defined by wrangler.jsonc `triggers.crons`, and
	//	  BlowActionSchedule must exactly match a configured cron expression.
	//	  "trigger" (start/stop) tasks are not supported on wasm.
	Blow(tasks ...BlowTask[Bindings])

	// Camp simulates a request without starting the server
	Camp(method, path string, opts ...CampOption) ICampResponse

	// just getter
	Router() IRouter[Bindings]
}
