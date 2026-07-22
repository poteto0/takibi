// Package cloudflareenv builds request-scoped Bindings from struct tags.
//
// A field tagged with `env:"NAME"` receives the value of the environment
// variable NAME — cloudflare.Getenv on wasm, os.Getenv on native. A field
// tagged with `cfbinding:"NAME"` receives the Cloudflare binding NAME on
// wasm and is left at its zero value on native, so the same main() builds
// for both targets.
//
//	type Bindings struct {
//		ApiKey string        `env:"API_KEY"`
//		Store  *kv.Namespace `cfbinding:"MY_KV"`
//	}
//
// takibi.New wires the resolver automatically when Bindings has any tagged
// field; call Resolver directly only to compose it with your own logic.
package cloudflareenv

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/poteto0/takibi/interfaces"
)

const (
	envTag       = "env"
	cfBindingTag = "cfbinding"
)

// fieldPlan describes one tagged field, resolved once at startup.
type fieldPlan struct {
	index int
	// name is the environment variable name, or the Cloudflare binding name
	// when binding is true.
	name    string
	binding bool
}

// buildPlan collects the tagged fields of typ. A non-struct typ has no
// taggable field and yields an empty plan.
func buildPlan(typ reflect.Type) ([]fieldPlan, error) {
	if typ.Kind() != reflect.Struct {
		return nil, nil
	}

	var plan []fieldPlan

	for i := range typ.NumField() {
		field := typ.Field(i)
		envName, hasEnv := field.Tag.Lookup(envTag)
		bindingName, hasBinding := field.Tag.Lookup(cfBindingTag)

		switch {
		case !hasEnv && !hasBinding:
			continue
		case hasEnv && hasBinding:
			return nil, fmt.Errorf("cloudflareenv: field %q: env and cfbinding tags are mutually exclusive", field.Name)
		case !field.IsExported():
			return nil, fmt.Errorf("cloudflareenv: field %q: tagged field must be exported", field.Name)
		case hasEnv && field.Type.Kind() != reflect.String:
			return nil, fmt.Errorf("cloudflareenv: field %q: env tag requires a string field, got %s", field.Name, field.Type)
		}

		if hasEnv {
			plan = append(plan, fieldPlan{index: i, name: envName})
			continue
		}
		plan = append(plan, fieldPlan{index: i, name: bindingName, binding: true})
	}

	return plan, nil
}

// Resolver returns a resolver that copies base and fills its tagged fields on
// every request. It returns a nil resolver when Bindings has no tagged field,
// so the caller can keep the app-wide Bindings as is.
func Resolver[Bindings any](base *Bindings) (interfaces.EnvResolverFunc[Bindings], error) {
	plan, err := buildPlan(reflect.TypeFor[Bindings]())
	if err != nil {
		return nil, err
	}
	if len(plan) == 0 {
		return nil, nil
	}

	if base == nil {
		base = new(Bindings)
	}

	return func(_ *http.Request) *Bindings {
		bindings := *base
		value := reflect.ValueOf(&bindings).Elem()

		for _, f := range plan {
			field := value.Field(f.index)
			if f.binding {
				assignBinding(field, f.name)
				continue
			}
			field.SetString(lookupEnv(f.name))
		}

		return &bindings
	}, nil
}
