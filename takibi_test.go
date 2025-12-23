package takibi

import (
	stdContext "context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func newNilApp() *takibi[any] {
	return New[any](nil).(*takibi[any])
}

func TestNewTakibi(t *testing.T) {
	type Bindings struct {
		Foo   string
		Greet func() string
	}

	t.Run("initializeContext", func(t *testing.T) {
		bindings := &Bindings{
			Foo: "bar",
			Greet: func() string {
				return "hello"
			},
		}

		app := New(bindings).(*takibi[Bindings])
		assert.Equal(t, "bar", app.Env().Foo)

		req1 := httptest.NewRequest(http.MethodGet, "/1", nil)
		w1 := httptest.NewRecorder()
		ctx1 := app.initializeContext(w1, req1)
		assert.Equal(t, req1, ctx1.Request())

		// Put back to cache
		app.cache.Put(ctx1)

		req2 := httptest.NewRequest(http.MethodGet, "/2", nil)
		w2 := httptest.NewRecorder()
		ctx2 := app.initializeContext(w2, req2)

		assert.Equal(t, ctx1, ctx2) // Should be same instance
		assert.Equal(t, req2, ctx2.Request())
		assert.Equal(t, w2, ctx2.Response())
	})

	t.Run("nil binding start", func(t *testing.T) {
		app := New[Bindings](nil)
		assert.Equal(t, app.Env().Foo, "")
	})
}

func TestTakibi_FireAndFinish(t *testing.T) {
	t.Run("can start & stop server", func(t *testing.T) {
		tests := []struct {
			name string
			port string
		}{
			{
				name: "port has port prefix",
				port: ":11111",
			},
			{
				name: "port doesn't have port prefix",
				port: "11111",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				app := newNilApp()

				errChan := make(chan error)
				go func() {
					err := app.Fire(test.port)

					// can start server
					if err != nil {
						errChan <- err
					}
				}()

				select {
				case <-time.After(10 * time.Millisecond):
					// can stop server
					assert.Nil(t, app.Finish(stdContext.Background()))
				case err := <-errChan:
					assert.Fail(t, fmt.Sprintf(
						"unexpected stop server before Finish because of %s",
						err.Error(),
					))
					return
				}
			})
		}
	})

	t.Run("use already created listener & fail to Finish", func(t *testing.T) {
		app := newNilApp()

		errChan := make(chan error)
		go func() {
			err := app.Fire("11111")

			if err != nil {
				errChan <- err
			}
		}()

		errChan2 := make(chan error)
		go func() {
			err := app.Fire("22222")

			if err != nil {
				errChan2 <- err
			}
		}()

		select {
		case <-time.After(10 * time.Millisecond):
			// can stop server
			assert.Error(t, app.Finish(stdContext.Background()))
		case err := <-errChan:
			assert.Fail(t, fmt.Sprintf(
				"unexpected stop server before Finish because of %s",
				err.Error(),
			))
			return
		case err := <-errChan2:
			assert.Fail(t, fmt.Sprintf(
				"unexpected stop server before Finish because of %s",
				err.Error(),
			))
			return
		}
	})

	t.Run("if duplicated port, return err", func(t *testing.T) {
		app := newNilApp()
		app2 := newNilApp()

		errChan := make(chan error)
		go func() {
			err := app.Fire("11111")

			if err != nil {
				errChan <- err
			}
		}()

		errChan2 := make(chan error)
		go func() {
			err := app2.Fire("11111")

			if err != nil {
				errChan2 <- err
			}
		}()

		select {
		case <-time.After(10 * time.Millisecond):
			// can stop server
			assert.Nil(t, app.Finish(stdContext.Background()))
			assert.Nil(t, app2.Finish(stdContext.Background()))
		// success on error
		case <-errChan:
			assert.Nil(t, app.Finish(stdContext.Background()))
			assert.Nil(t, app2.Finish(stdContext.Background()))
			return
		// success on error
		case <-errChan2:
			assert.Nil(t, app.Finish(stdContext.Background()))
			assert.Nil(t, app2.Finish(stdContext.Background()))
			return
		}
	})
}

func TestTakibi_addAllMethod(t *testing.T) {
	app := New[any](nil).(*takibi[any])
	assert.NotNil(t, app)

	t.Run("Get", func(t *testing.T) {
		assert.Nil(
			t,
			app.Get("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Post", func(t *testing.T) {
		assert.Nil(
			t,
			app.Post("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Put", func(t *testing.T) {
		assert.Nil(
			t,
			app.Put("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Patch", func(t *testing.T) {
		assert.Nil(
			t,
			app.Patch("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Delete", func(t *testing.T) {
		assert.Nil(
			t,
			app.Delete("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Head", func(t *testing.T) {
		assert.Nil(
			t,
			app.Head("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Options", func(t *testing.T) {
		assert.Nil(
			t,
			app.Options("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Trace", func(t *testing.T) {
		assert.Nil(
			t,
			app.Trace("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})

	t.Run("Connect", func(t *testing.T) {
		assert.Nil(
			t,
			app.Connect("/users", func(ctx interfaces.IContext[any]) error {
				return nil
			}),
		)
	})
}

func TestTakibi_ServeHTTP(t *testing.T) {
	app := New[any](nil)

	_ = app.Get("/hello", func(ctx interfaces.IContext[any]) error {
		ctx.Response().WriteHeader(http.StatusOK)
		_, err := ctx.Response().Write([]byte("hello world"))
		return err
	})

	t.Run("success request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello world", rec.Body.String())
	})

	t.Run("not found request on node is not exist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("not found request on handler is not exist", func(t *testing.T) {
		app.Get("/notfound", nil)

		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("internal server error on handler", func(t *testing.T) {
		app.Get("/error", func(ctx interfaces.IContext[any]) error {
			return fmt.Errorf("error")
		})

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("internal-server-error if error on error-handler", func(t *testing.T) {
		app.OnError(func(ctx interfaces.IContext[any], err error) error {
			return err
		})

		app.Get("/error", func(ctx interfaces.IContext[any]) error {
			return fmt.Errorf("error")
		})

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
