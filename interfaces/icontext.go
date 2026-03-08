package interfaces

import (
	"context"
	"net/http"
)

type IContext[Bindings any] interface {
	Env() *Bindings
	Req() IRequest
	Response() http.ResponseWriter
	Context() context.Context
	Reset(w http.ResponseWriter, r *http.Request)

	// Response
	Status(code int) IContext[Bindings]
	Text(text string) error
	Json(data any) error
	Redirect(url string) error
}
