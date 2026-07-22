//go:build wasm

package cloudflareenv

import (
	"reflect"
	"syscall/js"

	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/kv"
	"github.com/syumai/workers/cloudflare/r2"
)

func lookupEnv(name string) string {
	return cloudflare.Getenv(name)
}

var (
	kvNamespaceType = reflect.TypeFor[*kv.Namespace]()
	r2BucketType    = reflect.TypeFor[*r2.Bucket]()
	jsValueType     = reflect.TypeFor[js.Value]()
)

// assignBinding fills field with the Cloudflare binding named name.
// Supported field types are *kv.Namespace, *r2.Bucket and js.Value; any other
// type — or a binding missing from wrangler.jsonc — leaves the zero value.
func assignBinding(field reflect.Value, name string) {
	switch field.Type() {
	case kvNamespaceType:
		if namespace, err := kv.NewNamespace(name); err == nil {
			field.Set(reflect.ValueOf(namespace))
		}
	case r2BucketType:
		if bucket, err := r2.NewBucket(name); err == nil {
			field.Set(reflect.ValueOf(bucket))
		}
	case jsValueType:
		field.Set(reflect.ValueOf(cloudflare.GetBinding(name)))
	}
}
