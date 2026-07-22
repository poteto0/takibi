//go:build !wasm

package cloudflareenv

import (
	"os"
	"reflect"
)

func lookupEnv(name string) string {
	return os.Getenv(name)
}

// assignBinding is a no-op on native: Cloudflare bindings only exist on the
// Workers runtime, so a `cfbinding` field keeps its zero value.
func assignBinding(_ reflect.Value, _ string) {}
