//go:build wasm

package takibi

import (
	stdContext "context"
	"errors"
)

func (t *takibi[Bindings]) Fire(addr string) error {
	return errors.New("Fire is not supported in WASM")
}

func (t *takibi[Bindings]) Finish(ctx stdContext.Context) error {
	return errors.New("Finish is not supported in WASM")
}
