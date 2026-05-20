package main

import (
	"embed"
	"fmt"
	"html/template"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/syumai/workers"
)

//go:embed public/templates/*.html
var templatesFS embed.FS

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

	app.Renderer(
		map[string]*template.Template{
			"index": template.Must(
				template.New("index.html").ParseFS(templatesFS, "public/templates/index.html"),
			),
		},
	)

	app.OnError(func(ctx interfaces.IContext[Bindings], err error) error {
		return ctx.Status(500).Text("internal-server-error")
	})

	app.Get("/", func(ctx MyContext) error {
		return ctx.Render(
			&interfaces.RenderConfig{
				Key: "index",
			},
			nil,
		)
	})

	workers.Serve(app)
}
