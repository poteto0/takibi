package interfaces

import (
	"github.com/a-h/templ"
)

type RenderConfig struct {
	// template files key
	// set w/ renderer middleware
	Key string

	// templ component
	Component templ.Component

	// content type
	// default is text/html
	ContentType string
}

func (c *RenderConfig) IsTemplate() bool {
	return c.Component != nil
}
