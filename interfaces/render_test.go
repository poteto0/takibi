package interfaces_test

import (
	"html/template"
	"testing"

	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestRenderConfig_IsTemplate(t *testing.T) {
	t.Run("check IsTemplate method", func(t *testing.T) {
		config := &interfaces.RenderConfig{
			Key:      "test",
			Template: nil,
		}
		assert.False(t, config.IsTemplate())

		tmpl := template.Must(template.New("test").Parse("{{.}}"))
		config.Template = tmpl
		assert.True(t, config.IsTemplate())
	})
}
