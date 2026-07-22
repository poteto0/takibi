package main

import (
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"

	"cloudflare-worker/templates/pages"
)

// Bindings is filled per request: takibi injects the tagged fields for you.
// `env` reads a wrangler.jsonc var / secret on Workers and os.Getenv on a
// native build, so the same main() works for both targets. See the
// cloudflareenv package for the `cfbinding` tag (KV, R2, ...).
type Bindings struct {
	Greeting string `env:"GREETING"`
}

type MyContext = interfaces.IContext[Bindings]

func main() {
	app := takibi.New(&Bindings{})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: pages.Index(),
		})
	})

	app.Get("/greeting", func(ctx MyContext) error {
		return ctx.Text(ctx.Env().Greeting) // 100% string
	})

	app.Fire("")
}
