//go:build wasm

package takibi

import (
	"github.com/poteto0/takibi/interfaces"
)

func (t *takibi[Bindings]) Camp(method, path string, opts ...interfaces.CampOption) interfaces.ICampResponse {
	return nil // Or return a mock response that satisfies ICampResponse
}
