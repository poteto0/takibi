package main

import (
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi-docs/templates/layouts"
	"github.com/poteto0/takibi-docs/templates/pages"
	"github.com/poteto0/takibi/interfaces"
)

type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

func main() {
	app := takibi.New(&Bindings{})

	app.OnError(func(ctx MyContext, err error) error {
		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Home", "home", pages.Home()),
		})
	})

	app.Get("/getting-started", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Getting Started", "getting-started", pages.GettingStarted()),
		})
	})

	app.Get("/docs/type-safe-context", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Type-Safe Context", "type-safe-context", pages.TypeSafeContext()),
		})
	})

	app.Get("/docs/routing", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Routing", "routing", pages.Routing()),
		})
	})

	app.Get("/docs/redirect", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Redirect", "redirect", pages.Redirect()),
		})
	})

	app.Get("/docs/error-handling", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Error Handling", "error-handling", pages.ErrorHandling()),
		})
	})

	app.Get("/docs/request-body", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout("Request Body", "request-body", pages.RequestBody()),
		})
	})

	app.Fire("")
}
