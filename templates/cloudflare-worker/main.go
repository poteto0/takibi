package main

import (
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
)

// Bindings maps to Cloudflare Workers environment bindings (KV, secrets, etc.).
type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

func main() {
	app := takibi.New(&Bindings{})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Text("Hello from Cloudflare Workers!")
	})

	// Pass an empty string — address is ignored when built for WASM.
	app.Fire("")
}
