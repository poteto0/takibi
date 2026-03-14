package interfaces

import (
	"html/template"
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
	Redirect(url string) error

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
	//    if err := ctx.Steam(buf.Bytes()); err != nil {
	//        return err
	//    }
	//  }
	Steam(data []byte) error

	// render with template or template key in renderer map
	//
	//  config := &interfaces.RenderConfig{
	//      Template:    tmpl, // *template.Template
	//      ContentType: "text/html",
	//  }
	// or
	//  config := &interfaces.RenderConfig{
	//     TemplateKey: "test", // key in renderer map
	//      ContentType: "text/html",
	//  }
	// in handler
	//  ctx.Render(config, data)
	Render(config *RenderConfig, data any) error

	// Params
	Param() map[string]string
	ParamBy(key string) string
	SetParam(params map[string]string)

	RegisterRenderer(rendererMap map[string]*template.Template)
}
