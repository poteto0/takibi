package main

import (
	"fmt"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
)

// Bindings is resolved per request: takibi fills the tagged fields for you.
// `env` reads a wrangler.jsonc var / secret on Workers and os.Getenv on a
// native build, so the same main() works for both targets. Use `cfbinding`
// for KV, R2, D1 or Workers AI — those stay nil on a native build.
type Bindings struct {
	Foo   string `env:"FOO"`
	Greet func()
}

type MyContext = interfaces.IContext[Bindings]

func main() {
	// no resolver to register: takibi.New detects the `env` tag
	var bindings = &Bindings{
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
