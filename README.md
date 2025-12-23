LOGO

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

## Document

docs link
