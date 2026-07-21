# Hono-liked Web Backend Framework

This is Web-Framework for golang,
which can be hosted on cloudflare-workers w/ syumai-worker when wasm built.

- Type-safe context.

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

## Commands

```bash
just -l
```

## Rule

- develop with tdd
- always update docs(`/docs`), if api changed.
- build tags: `*_wasm.go` must carry `//go:build wasm`, `*_native.go` must carry `//go:build !wasm`.
  Go implies the constraint from the `_wasm` filename suffix but **not** from `_native`,
  so a `*_native.go` without the explicit tag is silently compiled into the wasm build too.
  Verify both targets with `just build-wasm` and `just lint-wasm`.

## docs

### Write golang code for docs

1. write golang-code to `docs-tool/assets/xxxgo.txt`
2. generate template

```bash
$ just gen-code assets/xxxgo.txt
```

3. refer method like

```templ
@code.Xxxgo()
```

4. generate go

```bash
$ just gen
```
