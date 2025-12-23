package main

import (
	"errors"
	"fmt"
)

type Bindings struct {
	Foo   string
	Greet func()
}

type MyContext = interfaces.IContext[Bindings]

var BadRequest = errors.New("bad request")

func main() {
	var bindings = &Bindings{
		Foo: "Bar",
		Greet: func() {
			fmt.Println("hello")
		},
	}

	app := takibi.New(bindings)

	app.OnError(func(ctx interfaces.IContext[Bindings], err error) error {
		if errors.Is(err, BadRequest) {
			return ctx.Status(400).Text("bad-request")
		}

		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/hello", func(ctx MyContext) error {
		fmt.Println(ctx.Env().Foo)
		ctx.Env().Greet()
		return nil
	})

	app.Get("/error", func(ctx MyContext) error {
		return BadRequest
	})

	if err := app.Fire(":8000"); err != nil {
		panic(err)
	}
}
