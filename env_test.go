package takibi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestTakibi_OnEnv(t *testing.T) {
	type Bindings struct{ Foo string }

	t.Run("resolver is called per request", func(t *testing.T) {
		app := New[Bindings](nil)
		called := 0
		app.OnEnv(func(r *http.Request) *Bindings {
			called++
			return &Bindings{Foo: r.URL.Path}
		})

		seen := make([]string, 0, 2)
		_ = app.Get("/a", func(ctx interfaces.IContext[Bindings]) error {
			seen = append(seen, ctx.Env().Foo)
			return ctx.Text("ok")
		})
		_ = app.Get("/b", func(ctx interfaces.IContext[Bindings]) error {
			seen = append(seen, ctx.Env().Foo)
			return ctx.Text("ok")
		})

		// two sequential requests go through the sync.Pool reuse path
		app.Camp(http.MethodGet, "/a")
		app.Camp(http.MethodGet, "/b")

		assert.Equal(t, 2, called)
		assert.Equal(t, []string{"/a", "/b"}, seen)
	})

	t.Run("without resolver the app env is kept", func(t *testing.T) {
		app := New(&Bindings{Foo: "fixed"})

		seen := make([]string, 0, 2)
		_ = app.Get("/", func(ctx interfaces.IContext[Bindings]) error {
			seen = append(seen, ctx.Env().Foo)
			return ctx.Text("ok")
		})

		app.Camp(http.MethodGet, "/")
		app.Camp(http.MethodGet, "/")

		assert.Equal(t, []string{"fixed", "fixed"}, seen)
	})

	t.Run("pooled context does not leak the previous env", func(t *testing.T) {
		app := New(&Bindings{Foo: "fixed"}).(*takibi[Bindings])

		r1 := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx1 := app.initializeContext(httptest.NewRecorder(), r1)
		ctx1.SetEnv(&Bindings{Foo: "leaked"})
		app.cache.Put(ctx1)

		r2 := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx2 := app.initializeContext(httptest.NewRecorder(), r2)

		assert.Equal(t, "fixed", ctx2.Env().Foo)
	})
}

func TestTakibi_TaggedBindings(t *testing.T) {
	t.Run("tagged fields are injected per request without OnEnv", func(t *testing.T) {
		type Bindings struct {
			ApiKey string `env:"TAKIBI_TEST_API_KEY"`
			Prefix string
		}

		app := New(&Bindings{Prefix: "kept"})
		seen := make([]string, 0, 2)
		_ = app.Get("/", func(ctx interfaces.IContext[Bindings]) error {
			seen = append(seen, ctx.Env().Prefix+":"+ctx.Env().ApiKey)
			return ctx.Text("ok")
		})

		t.Setenv("TAKIBI_TEST_API_KEY", "first")
		app.Camp(http.MethodGet, "/")
		t.Setenv("TAKIBI_TEST_API_KEY", "second")
		app.Camp(http.MethodGet, "/")

		assert.Equal(t, []string{"kept:first", "kept:second"}, seen)
	})

	t.Run("OnEnv overrides the tag resolver", func(t *testing.T) {
		type Bindings struct {
			ApiKey string `env:"TAKIBI_TEST_API_KEY"`
		}

		app := New[Bindings](nil)
		app.OnEnv(func(_ *http.Request) *Bindings {
			return &Bindings{ApiKey: "custom"}
		})
		seen := ""
		_ = app.Get("/", func(ctx interfaces.IContext[Bindings]) error {
			seen = ctx.Env().ApiKey
			return ctx.Text("ok")
		})

		t.Setenv("TAKIBI_TEST_API_KEY", "from-env")
		app.Camp(http.MethodGet, "/")

		assert.Equal(t, "custom", seen)
	})

	t.Run("an invalid tag panics at construction", func(t *testing.T) {
		type Bindings struct {
			Port int `env:"PORT"`
		}

		assert.Panics(t, func() { New[Bindings](nil) })
	})
}

func TestTakibi_NewTaskContext(t *testing.T) {
	type Bindings struct {
		ApiKey string `env:"TAKIBI_TEST_API_KEY"`
	}

	t.Setenv("TAKIBI_TEST_API_KEY", "resolved")
	app := New[Bindings](nil).(*takibi[Bindings])

	c := app.newTaskContext(httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "resolved", c.Env().ApiKey)
}
