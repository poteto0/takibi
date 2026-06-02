package main

import (
	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
)

type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

func main() {
	app := takibi.New(&Bindings{})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Render(&interfaces.RenderConfig{
			Component: Index(),
		})
	})

	app.Fire("")
}
