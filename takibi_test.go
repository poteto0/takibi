package takibi

import (
	stdContext "context"
	"fmt"
	"html/template"
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
		assert.Equal(t, req1, ctx1.Req().Raw())

		// Put back to cache
		app.cache.Put(ctx1)

		req2 := httptest.NewRequest(http.MethodGet, "/2", nil)
		w2 := httptest.NewRecorder()
		ctx2 := app.initializeContext(w2, req2)

		assert.Equal(t, ctx1, ctx2) // Should be same instance
		assert.Equal(t, req2, ctx2.Req().Raw())
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

var handler = func(c interfaces.IContext[any]) error {
	return nil
}

func TestTakibi_Router(t *testing.T) {
	t.Run("can add route w/ base path", func(t *testing.T) {
		// Arrange
		app1 := newNilApp()
		app2 := newNilApp()

		app1.Get("/users", handler)
		app2.Get("/hello", handler)

		// Act
		err := app1.Route("/greet", app2)

		// Assert
		assert.Nil(t, err)

		resp := app1.Camp("GET", "/greet/hello")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("return error if duplicate when Route", func(t *testing.T) {
		// Arrange
		app1 := newNilApp()
		app2 := newNilApp()

		app1.Get("/greet/users", handler)
		app2.Get("/users", handler)

		// Act
		err := app1.Route("/greet", app2)

		// Assert
		assert.Error(t, err)
	})
}

func TestTakibi_addAllMethod(t *testing.T) {
	app := New[any](nil).(*takibi[any])
	assert.NotNil(t, app)

	t.Run("Get", func(t *testing.T) {
		err := app.Get("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("GET", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Post", func(t *testing.T) {
		err := app.Post("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("POST", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Put", func(t *testing.T) {
		err := app.Put("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("PUT", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Patch", func(t *testing.T) {
		err := app.Patch("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("PATCH", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Delete", func(t *testing.T) {
		err := app.Delete("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("DELETE", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Head", func(t *testing.T) {
		err := app.Head("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("HEAD", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Options", func(t *testing.T) {
		err := app.Options("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("OPTIONS", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Trace", func(t *testing.T) {
		err := app.Trace("/users", handler)
		assert.Nil(
			t,
			err,
		)

		resp := app.Camp("TRACE", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("Connect", func(t *testing.T) {
		err := app.Connect("/users", handler)
		assert.Nil(t, err)

		resp := app.Camp("CONNECT", "/users")
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("All", func(t *testing.T) {
		err := app.All("/all", handler)
		assert.Nil(t, err)

		methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT"}
		for _, method := range methods {
			resp := app.Camp(method, "/all")
			assert.Equal(t, http.StatusOK, resp.StatusCode())
		}

		// error case
		app.Get("/get", handler)
		err = app.All("/get", handler)
		assert.Error(t, err)

		app.Post("/post", handler)
		err = app.All("/post", handler)
		assert.Error(t, err)

		app.Put("/put", handler)
		err = app.All("/put", handler)
		assert.Error(t, err)

		app.Patch("/patch", handler)
		err = app.All("/patch", handler)
		assert.Error(t, err)

		app.Delete("/delete", handler)
		err = app.All("/delete", handler)
		assert.Error(t, err)

		app.Head("/head", handler)
		err = app.All("/head", handler)
		assert.Error(t, err)

		app.Options("/options", handler)
		err = app.All("/options", handler)
		assert.Error(t, err)

		app.Trace("/trace", handler)
		err = app.All("/trace", handler)
		assert.Error(t, err)

		app.Connect("/connect", handler)
		err = app.All("/connect", handler)
		assert.Error(t, err)
	})

	t.Run("On", func(t *testing.T) {
		t.Run("register all", func(t *testing.T) {
			err := app.On(
				[]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodPatch,
					http.MethodDelete,
					http.MethodHead,
					http.MethodOptions,
					http.MethodTrace,
					http.MethodConnect,
				},
				[]string{
					"/on1",
					"/on2",
				},
				handler,
			)
			assert.Nil(t, err)

			resp := app.Camp("GET", "/on1")
			assert.Equal(t, http.StatusOK, resp.StatusCode())

			resp = app.Camp("POST", "/on2")
			assert.Equal(t, http.StatusOK, resp.StatusCode())
		})

		t.Run("invalid method error", func(t *testing.T) {
			err := app.On([]string{"invalid"}, []string{"/invalid"}, handler)
			assert.Error(t, err)
		})

		t.Run("inner error", func(t *testing.T) {
			err := app.On([]string{http.MethodGet}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodPost}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodPut}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodPatch}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodDelete}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodHead}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodOptions}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodTrace}, []string{"/on1"}, handler)
			assert.Error(t, err)

			err = app.On([]string{http.MethodConnect}, []string{"/on1"}, handler)
			assert.Error(t, err)
		})
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

	t.Run("path param", func(t *testing.T) {
		app := New[any](nil)
		app.Get("/users/:id", func(ctx interfaces.IContext[any]) error {
			id := ctx.ParamBy("id")
			return ctx.Text("user " + id)
		})

		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "user 123", rec.Body.String())
	})
}

func TestTakibi_Renderer(t *testing.T) {
	app := New[any](nil).(*takibi[any])

	tmpl := template.Must(template.New("test").Parse("Hello {{.Name}}"))
	rendererMap := map[string]*template.Template{
		"test": tmpl,
	}
	app.Renderer(rendererMap)

	assert.Equal(t, rendererMap, app.rendererMap)
}
