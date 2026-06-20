package main

import (
	"github.com/a-h/templ"
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi-docs/templates/layouts"
	"github.com/poteto0/takibi-docs/templates/pages"
	"github.com/poteto0/takibi/interfaces"
)

type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

func docPage(title, active string, component templ.Component) func(ctx MyContext) error {
	return func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: layouts.Layout(title, active, component),
		})
	}
}

func main() {
	app := takibi.New(&Bindings{})

	app.OnError(func(ctx MyContext, err error) error {
		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/", docPage("Home", "home", pages.Home()))
	app.Get("/getting-started", docPage("Getting Started", "getting-started", pages.GettingStarted()))
	app.Get("/getting-started/cloudflare-workers", docPage("Cloudflare Workers", "cloudflare-workers", pages.CloudflareWorker()))

	app.Get("/docs/type-safe-context", docPage("Type-Safe Context", "type-safe-context", pages.TypeSafeContext()))
	app.Get("/docs/routing", docPage("Routing", "routing", pages.Routing()))
	app.Get("/docs/sub-routing", docPage("Sub-Routing", "sub-routing", pages.SubRouting()))

	app.Get("/docs/middleware", docPage("Middleware", "middleware", pages.Middleware()))
	app.Get("/docs/redirect", docPage("Redirect", "redirect", pages.Redirect()))
	app.Get("/docs/error-handling", docPage("Error Handling", "error-handling", pages.ErrorHandling()))
	app.Get("/docs/request-body", docPage("Request Body", "request-body", pages.RequestBody()))
	app.Get("/docs/validator", docPage("Validator", "validator", pages.Validator()))
	app.Get("/docs/cookie", docPage("Cookie", "cookie", pages.Cookie()))

	app.Get("/docs/factory", docPage("Factory Helpers", "factory", pages.Factory()))
	app.Get("/docs/testing", docPage("Testing", "testing", pages.Testing()))
	app.Get("/docs/blow", docPage("Background Tasks", "blow", pages.Blow()))

	app.Fire("")
}
