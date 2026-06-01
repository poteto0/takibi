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
