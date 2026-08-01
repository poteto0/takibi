//go:build wasm

package cloudflareenv

import (
	"database/sql"
	"reflect"
	"syscall/js"

	"github.com/poteto0/takibi/cfextensions"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/kv"
	"github.com/syumai/workers/cloudflare/r2"

	_ "github.com/syumai/workers/cloudflare/d1" // register driver
)

func lookupEnv(name string) string {
	return cloudflare.Getenv(name)
}

var (
	d1Type          = reflect.TypeFor[*sql.DB]()
	kvNamespaceType = reflect.TypeFor[*kv.Namespace]()
	r2BucketType    = reflect.TypeFor[*r2.Bucket]()
	workersAiType   = reflect.TypeFor[*cfextensions.WorkersAi]()
	jsValueType     = reflect.TypeFor[js.Value]()
)

// assignBinding fills field with the Cloudflare binding named name.
// Supported field types are *kv.Namespace, *r2.Bucket, *sql.DB (D1),
// *cfextensions.WorkersAi and js.Value; any other type — or a binding missing
// from wrangler.jsonc — leaves the zero value.
func assignBinding(field reflect.Value, name string) {
	switch field.Type() {
	case kvNamespaceType:
		if namespace, err := kv.NewNamespace(name); err == nil {
			field.Set(reflect.ValueOf(namespace))
		}
	case d1Type:
		if db, err := sql.Open("d1", name); err == nil {
			field.Set(reflect.ValueOf(db))
		}
	case r2BucketType:
		if bucket, err := r2.NewBucket(name); err == nil {
			field.Set(reflect.ValueOf(bucket))
		}
	case workersAiType:
		field.Set(reflect.ValueOf(&cfextensions.WorkersAi{
			Instance: cloudflare.GetBinding(name),
		}))
	case jsValueType:
		field.Set(reflect.ValueOf(cloudflare.GetBinding(name)))
	}
}
