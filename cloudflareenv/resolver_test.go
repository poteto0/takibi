package cloudflareenv

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type unsupportedBinding struct{ Ptr *int }

func TestBuildPlan(t *testing.T) {
	tests := []struct {
		name    string
		typ     any
		want    []fieldPlan
		wantErr string
	}{
		{
			name: "no tag",
			typ:  struct{ Foo string }{},
			want: nil,
		},
		{
			name: "non struct type",
			typ:  "not a struct",
			want: nil,
		},
		{
			name: "env and cfbinding tags",
			typ: struct {
				ApiKey string `env:"API_KEY"`
				Plain  string
				Store  unsupportedBinding `cfbinding:"MY_KV"`
			}{},
			want: []fieldPlan{
				{index: 0, name: "API_KEY"},
				{index: 2, name: "MY_KV", binding: true},
			},
		},
		{
			name: "env tag on non-string field",
			typ: struct {
				Port int `env:"PORT"`
			}{},
			wantErr: `cloudflareenv: field "Port": env tag requires a string field, got int`,
		},
		{
			name: "both tags on one field",
			typ: struct {
				Foo string `env:"FOO" cfbinding:"FOO"`
			}{},
			wantErr: `cloudflareenv: field "Foo": env and cfbinding tags are mutually exclusive`,
		},
		{
			name: "unexported tagged field",
			typ: struct {
				foo string `env:"FOO"`
			}{},
			wantErr: `cloudflareenv: field "foo": tagged field must be exported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPlan(reflect.TypeOf(tt.typ))
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolver(t *testing.T) {
	t.Run("returns nil when Bindings has no tag", func(t *testing.T) {
		type Bindings struct{ Foo string }

		resolver, err := Resolver(&Bindings{Foo: "bar"})

		assert.NoError(t, err)
		assert.Nil(t, resolver)
	})

	t.Run("returns the plan error", func(t *testing.T) {
		type Bindings struct {
			Port int `env:"PORT"`
		}

		resolver, err := Resolver(&Bindings{})

		assert.Error(t, err)
		assert.Nil(t, resolver)
	})

	t.Run("fills tagged fields per request and keeps untagged ones", func(t *testing.T) {
		type Bindings struct {
			ApiKey string `env:"API_KEY"`
			Greet  func() string
		}

		base := &Bindings{Greet: func() string { return "hello" }}
		resolver, err := Resolver(base)
		assert.NoError(t, err)

		t.Setenv("API_KEY", "first")
		got1 := resolver(httptest.NewRequest(http.MethodGet, "/", nil))
		t.Setenv("API_KEY", "second")
		got2 := resolver(httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, "first", got1.ApiKey)
		assert.Equal(t, "second", got2.ApiKey)
		assert.NotSame(t, got1, got2)
		assert.Equal(t, "", base.ApiKey)
		assert.Equal(t, "hello", got1.Greet())
	})

	t.Run("cfbinding fields stay zero on native", func(t *testing.T) {
		type Bindings struct {
			Store unsupportedBinding `cfbinding:"MY_KV"`
		}

		resolver, err := Resolver(&Bindings{})
		assert.NoError(t, err)

		got := resolver(httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Nil(t, got.Store.Ptr)
	})

	t.Run("accepts a nil base", func(t *testing.T) {
		type Bindings struct {
			ApiKey string `env:"API_KEY"`
		}

		resolver, err := Resolver[Bindings](nil)
		assert.NoError(t, err)

		t.Setenv("API_KEY", "value")
		assert.Equal(t, "value", resolver(httptest.NewRequest(http.MethodGet, "/", nil)).ApiKey)
	})
}
