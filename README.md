<img src="./imgs/logo.svg" alt="takibi" width=500>

## Hono Inspired Type-Safe Context Web Framework

### 🛡️ Type-Safe Context

No need to worry about any type of store anymore.

```go
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
	app.Get("/hello", func(ctx MyContext) error {
		fmt.Println(ctx.Env().Foo) // 100% string
		ctx.Env().Greet() // 100% func()
		return nil
	})

	if err := app.Fire(":8000"); err != nil {
		panic(err)
	}
}
```

## Safe Redirect

`ctx.Redirect()` only accepts relative paths. For redirecting to external hosts, use `ctx.RedirectExternal()` with an explicit allowlist.

```go
type Bindings struct {
    AllowedRedirectHosts []string
}

type MyContext = interfaces.IContext[Bindings]

func main() {
    bindings := &Bindings{
        AllowedRedirectHosts: []string{"auth.example.com"},
    }

    app := takibi.New(bindings)

    // relative redirect — always safe
    app.Get("/dashboard", func(ctx MyContext) error {
        return ctx.Redirect("/home")
    })

    // external redirect — host validated against allowlist
    app.Get("/oauth/callback", func(ctx MyContext) error {
        next := ctx.Req().QueryBy("next")
        return ctx.RedirectExternal(next, ctx.Env().AllowedRedirectHosts)
    })
}
```

## Error Handling

By default, takibi responds with a generic `"Internal Server Error"` message for unhandled errors — raw error details are never exposed to clients. Use `OnError` to customize the behavior:

```go
app.OnError(func(ctx interfaces.IContext[Bindings], err error) error {
    // log err internally if needed
    return ctx.Status(http.StatusInternalServerError).Text("something went wrong")
})
```

## Request Body Size Limit

`ctx.Req().Unmarshall()` enforces a default limit of **10 MiB** per request body to prevent DoS via large payloads. Configure it per app via `WithMaxBodyBytes`:

```go
app := takibi.New(bindings, takibi.WithMaxBodyBytes[Bindings](4<<20)) // 4 MiB

app.Post("/upload", func(ctx MyContext) error {
    var payload MyPayload
    if err := ctx.Req().Unmarshall(&payload); err != nil {
        var maxErr *http.MaxBytesError
        if errors.As(err, &maxErr) {
            return ctx.Status(http.StatusRequestEntityTooLarge).Text("request body too large")
        }
        return err
    }
    return ctx.Text("ok")
})
```

The constant `thttp.DefaultMaxBodyBytes` (= `10 << 20`) is also available if you need to reference the default.

## Document

docs link
