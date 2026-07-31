//go:build wasm

package cfextensions

import (
	"fmt"
	"syscall/js"

	"github.com/poteto0/takibi/ffi"
	jsutil "github.com/syumai/go-jsutil"
)

// WorkersAi is the Workers AI binding. Declare it in your Bindings struct and
// the resolver fills Instance on every request:
//
//	type Bindings struct {
//		AI *cfextensions.WorkersAi `cfbinding:"AI"`
//	}
type WorkersAi struct {
	// Model is the default model, used when RequestOption.Model is empty.
	Model string
	// Instance is the JavaScript `Ai` object handed over by the runtime.
	Instance js.Value
}

// run calls `AI.run(model, payload)` and returns the JSON-encoded result.
func (w *WorkersAi) run(opt RequestOption) ([]byte, error) {
	if w.Instance.IsUndefined() || w.Instance.IsNull() {
		return nil, ErrAiUnavailable
	}

	payload, err := ffi.ToJSValue(opt)
	if err != nil {
		return nil, fmt.Errorf("cfextensions: failed to build payload: %w", err)
	}

	res, err := jsutil.AwaitPromise(w.Instance.Call("run", opt.Model, payload))
	if err != nil {
		return nil, fmt.Errorf("cfextensions: run failed: %w", err)
	}

	raw, err := ffi.FromJSValue(res)
	if err != nil {
		return nil, fmt.Errorf("cfextensions: failed to read response: %w", err)
	}

	return raw, nil
}
