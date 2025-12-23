package helper

import (
	"github.com/a-h/templ"
	"github.com/poteto0/takibi/interfaces"
)

func Render[TemplateArg any, Bindings any](
	ctx interfaces.IContext[Bindings],
	componentFunc func(args TemplateArg) templ.Component,
	args TemplateArg,
) error {
	ctx.Response().Header().Set("Content-Type", "text/html")
	component := componentFunc(args)
	return component.Render(ctx.Request().Context(), ctx.Response())
}
