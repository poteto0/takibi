package interfaces

import (
	"net/http"
)

type IContext[Bindings any] interface {
	Env() *Bindings
	Req() IRequest
	Response() http.ResponseWriter
	Reset(w http.ResponseWriter, r *http.Request)

	// Response
	Status(code int) IContext[Bindings]
	Text(text string) error
	Bytes(data []byte) error
	Json(data any) error
	// Redirect sends a 302 response to a relative path.
	// Returns an error if url is an absolute URL or protocol-relative URL.
	// Use RedirectExternal for redirecting to external hosts.
	Redirect(path string) error

	// RedirectExternal sends a 302 response to an absolute URL.
	// The host of url must appear in allowedHosts; returns an error otherwise.
	RedirectExternal(url string, allowedHosts []string) error

	// don't write header and status code
	// so you can use for streaming response
	//  c.Response().Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	//
	//  for {
	//    var buf bytes.Buffer
	//    buf.WriteString("--frame\r\n")
	//    buf.WriteString("Content-Type: image/jpeg\r\n\r\n")
	//    buf.Write(data) // data is []byte of jpeg image
	//    buf.WriteString("\r\n")
	//
	//    if err := ctx.Stream(buf.Bytes()); err != nil {
	//        return err
	//    }
	//  }
	Stream(data []byte) error

	// render with component
	//
	//  config := &interfaces.RenderConfig{
	//      Component:   component, // templ.Component
	//      ContentType: "text/html",
	//  }
	// in handler
	//  ctx.Render(config)
	Render(config *RenderConfig) error

	// Params
	Param() map[string]string
	ParamBy(key string) string
	SetParam(params map[string]string)
}
