//go:build wasm

package takibi

type engine[Bindings any] struct{}

func (t *takibi[Bindings]) initEngine() {}
