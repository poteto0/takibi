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

	app.OnError(func(ctx MyContext, err error) error {
		return ctx.Status(500).Text("internal-server-error")
	})

	app.OnBlowError(func(ctx MyContext, err error) {
		fmt.Println("scheduled task failed:", err.Error())
	})

	app.Get("/hello", func(ctx MyContext) error {
		fmt.Println(ctx.Env().Foo)
		ctx.Env().Greet()
		return ctx.Json(map[string]string{
			"message": "hello world",
		})
	})

	// Scheduled task dispatched by Cloudflare Cron Triggers.
	// BlowActionSchedule must exactly match an entry in wrangler.jsonc
	// `triggers.crons` (standard 5-field cron, no seconds).
	app.Blow(interfaces.BlowTask[Bindings]{
		BlowActionTag:      "schedule",
		BlowActionSchedule: "*/5 * * * *",
		BlowAction: func(ctx MyContext) error {
			fmt.Println("cron fired:", ctx.Env().Foo)
			return nil
		},
	})

	// Fire wires HTTP handlers and Cron Triggers, then blocks until done.
	// The addr argument is ignored on WASM.
	if err := app.Fire(""); err != nil {
		panic(err)
	}
}
