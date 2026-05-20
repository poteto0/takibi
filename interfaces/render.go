package interfaces

type RenderConfig struct {
	// template files key
	// set w/ renderer middleware
	Key string

	// template
	Template any

	// content type
	// default is text/html
	ContentType string
}

func (c *RenderConfig) IsTemplate() bool {
	return c.Template != nil
}
