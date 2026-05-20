package interfaces_test

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestRenderConfig_IsTemplate(t *testing.T) {
	t.Run("check IsTemplate method", func(t *testing.T) {
		config := &interfaces.RenderConfig{
			Key:       "test",
			Component: nil,
		}
		assert.False(t, config.IsTemplate())

		component := templ.NopComponent
		config.Component = component
		assert.True(t, config.IsTemplate())
	})
}
