package main

import (
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi-docs/public/templates"
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
			Component: templates.Layout("Home", "home", templates.Home()),
		})
	})

	app.Get("/getting-started", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Getting Started", "getting-started", templates.GettingStarted()),
		})
	})

	app.Get("/docs/type-safe-context", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Type-Safe Context", "type-safe-context", templates.TypeSafeContext()),
		})
	})

	app.Get("/docs/routing", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Routing", "routing", templates.Routing()),
		})
	})

	app.Get("/docs/redirect", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Redirect", "redirect", templates.Redirect()),
		})
	})

	app.Get("/docs/error-handling", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Error Handling", "error-handling", templates.ErrorHandling()),
		})
	})

	app.Get("/docs/request-body", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: templates.Layout("Request Body", "request-body", templates.RequestBody()),
		})
	})

	app.Fire("")
}
