//go:build !wasm

package cfextensions

// WorkersAi is the Workers AI binding. It carries no JavaScript instance on a
// native build — the type exists so the same Bindings struct compiles for both
// targets — and Request always fails with ErrAiUnavailable.
type WorkersAi struct {
	// Model is the default model, used when RequestOption.Model is empty.
	Model string
}

func (w *WorkersAi) run(_ RequestOption) ([]byte, error) {
	return nil, ErrAiUnavailable
}
