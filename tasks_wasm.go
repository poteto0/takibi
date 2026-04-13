//go:build wasm

package takibi

import (
	stdContext "context"
)

func (t *takibi[Bindings]) startTasks() {}

func (t *takibi[Bindings]) stopTasks(ctx stdContext.Context) {}
