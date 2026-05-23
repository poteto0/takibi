package interfaces

import (
	"github.com/a-h/templ"
)

type RenderConfig struct {
	// templ component
	Component templ.Component

	// content type
	// default is text/html
	ContentType string
}
