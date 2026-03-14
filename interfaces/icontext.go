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
	Bytes(data []byte) error
	Json(data any) error
	Redirect(url string) error
	// don't write header and status code, just write data
	// boundary is fixed to "frame"
	//
	//  var buf bytes.Buffer
	//  buf.WriteString("--frame\r\n")
	//  buf.WriteString("Content-Type: image/jpeg\r\n\r\n")
	//  buf.Write(data) // data is []byte of jpeg image
	//  buf.WriteString("\r\n")
	//
	//  if err := ctx.Steam(buf.Bytes()); err != nil {
	//      return err
	//  }
	Steam(data []byte) error

	// Params
	Param() map[string]string
	ParamBy(key string) string
	SetParam(params map[string]string)
}
