package main

import (
	"fmt"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
)

type Bindings struct {
	Foo   string
	Greet func()
}

type MyContext = interfaces.IContext[Bindings]

func main() {
	var bindings = &Bindings{
		Foo: "Bar",
		Greet: func() {
			fmt.Println("hello")
		},
	}

	app := takibi.New(bindings)

	app.OnError(func(ctx interfaces.IContext[Bindings], err error) error {
		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Render(
			&interfaces.RenderConfig{
				Component: Hello("Takibi"),
			},
			nil,
		)
	})

	app.Fire("")
}
