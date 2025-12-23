package main

import (
	"errors"
	"fmt"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
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
			return ctx.Status(400).Json(
				map[string]string{
					"message": "error",
				},
			)
		}

		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/hello", func(ctx MyContext) error {
		fmt.Println(ctx.Env().Foo)
		ctx.Env().Greet()
		return ctx.Json(map[string]string{
			"message": "hello world",
		})
	})

	app.Get("/error", func(ctx MyContext) error {
		return BadRequest
	})

	if err := app.Fire(":8000"); err != nil {
		panic(err)
	}
}
