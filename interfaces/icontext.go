package interfaces

import (
	"context"
	"net/http"
)

type IContext[Bindings any] interface {
	Env() *Bindings
	Request() *http.Request
	Response() http.ResponseWriter
	Context() context.Context
	Reset(w http.ResponseWriter, r *http.Request)

	Status(code int) IContext[Bindings]
	Text(text string) error
	Json(data any) error
	Redirect(url string) error
}
