//go:build wasm

package ffi

import (
	"encoding/json"
	"errors"
	"syscall/js"
)

// ToJSValue marshals v to JSON and revives it as a JavaScript value, so a Go
// struct can be handed to a Workers binding as a plain JS object.
func ToJSValue(v any) (js.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return js.Undefined(), err
	}
	return js.Global().Get("JSON").Call("parse", string(b)), nil
}

// FromJSValue serializes a JavaScript value back to JSON bytes, ready for
// json.Unmarshal. It fails when v is undefined, null, or not serializable —
// JSON.stringify returns undefined for a function or a lone symbol.
func FromJSValue(v js.Value) ([]byte, error) {
	if v.IsUndefined() || v.IsNull() {
		return nil, errors.New("ffi: value is undefined or null")
	}

	s := js.Global().Get("JSON").Call("stringify", v)
	if s.Type() != js.TypeString {
		return nil, errors.New("ffi: value is not JSON serializable")
	}
	return []byte(s.String()), nil
}
