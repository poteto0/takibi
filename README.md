<img src="./imgs/logo.svg" alt="takibi" width=500>

## Hono Inspired Type-Safe Context Web Framework

### 🛡️ Auto-Resolved, Type-Safe Bindings

Tag the fields of your `Bindings` struct and takibi fills them for every request — no store, no casts, no env plumbing.

```go
type Bindings struct {
	// wasm: cloudflare.Getenv("API_KEY") / native: os.Getenv("API_KEY")
	ApiKey string `env:"API_KEY"`
	// wasm: the MY_KV binding / native: left nil
	Store *kv.Namespace `cfbinding:"MY_KV"`
	// untagged: keeps the value given to takibi.New
	Version string
}

type MyContext = interfaces.IContext[Bindings]

func main() {
	// no resolver to register: takibi.New detects the tags
	app := takibi.New(&Bindings{Version: "v1"})

	app.Get("/secret", func(ctx MyContext) error {
		// 100% string, resolved for this request only
		return ctx.Text(ctx.Env().ApiKey)
	})

	if err := app.Fire(":8000"); err != nil {
		panic(err)
	}
}
```

| Tag | Field type | Workers (wasm) | Native |
| --- | --- | --- | --- |
| `env:"NAME"` | `string` only | `cloudflare.Getenv("NAME")` | `os.Getenv("NAME")` |
| `cfbinding:"NAME"` | `*kv.Namespace`, `*r2.Bucket`, `*sql.DB` (D1), `*cfextensions.WorkersAi`, `js.Value` | the Cloudflare binding `NAME` | zero value |

Both tags on one field, an unexported tagged field, or `env` on a non-string field panics at `takibi.New` instead of failing on the first request. A `cfbinding` field is simply zero on a native build, so the same `main()` builds for both targets.

### 🧩 Plain Bindings

Nothing forces you to use tags — any typed value works, and `ctx.Env()` returns it as is.

```go
type Bindings struct {
	Foo   string
	Greet func()
}

app := takibi.New(&Bindings{
	Foo:   "Bar",
	Greet: func() { fmt.Println("hello") },
})

app.Get("/hello", func(ctx MyContext) error {
	fmt.Println(ctx.Env().Foo) // 100% string
	ctx.Env().Greet()          // 100% func()
	return nil
})
```

Without any tag and without `OnEnv`, every request shares the single `Bindings` pointer passed to `takibi.New`, so writing to `ctx.Env()` from a handler is a data race across concurrent requests.

### 🔧 Custom Resolvers

For values derived from the request itself, register your own resolver with `OnEnv`. It runs once per request and replaces the tag resolver entirely.

```go
app.OnEnv(func(r *http.Request) *Bindings {
	return &Bindings{
		ApiKey:    os.Getenv("API_KEY"),
		RequestID: r.Header.Get("X-Request-Id"),
	}
})
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

`ctx.Req().Unmarshall()` enforces a default limit of **10 MiB** per request body to prevent DoS via large payloads. Use `NewWithOption` to configure it:

```go
app := takibi.NewWithOption(bindings, takibi.TakibiOption{
    MaxBodyBytes: 4 << 20, // 4 MiB
})

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

`takibi.New` uses the default limit (`constants.DefaultMaxBodyBytes` = 10 MiB).

`Unmarshall` requires a JSON request body. The `Content-Type` is matched on its media type, so values carrying parameters such as `application/json; charset=utf-8` are accepted.

## Signed Cookies

`cookie.SetSignedCookie` and `cookie.GetSignedCookie` HMAC-sign cookie values using `gorilla/securecookie`. The `secret` must be **at least 32 bytes**; shorter secrets are rejected and the functions return `false`/`nil, false` immediately.

```go
import "github.com/poteto0/takibi/cookie"

// secret must be >= 32 bytes
secret := "my-32-byte-or-longer-secret-key!!"

// set
ok := cookie.SetSignedCookie[Bindings](ctx, "session", userID, secret, nil)

// get (returns decoded value; false if missing, tampered, or wrong secret)
c, ok := cookie.GetSignedCookie[Bindings](ctx, "session", secret, nil)
```

Use `constants.MinSignedCookieSecretLen` (32) as the documented minimum when generating secrets.

## Document

docs link
