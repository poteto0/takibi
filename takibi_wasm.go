//go:build wasm

package takibi

import (
	stdContext "context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/router"
	"github.com/poteto0/takibi/thttp"
	"github.com/syumai/workers"
)

type takibi[Bindings any] struct {
	env          *Bindings
	cache        sync.Pool
	router       interfaces.IRouter[Bindings]
	errorHandler interfaces.ErrorHandlerFunc[Bindings]
	maxBodyBytes int64

	ctx    stdContext.Context
	cancel stdContext.CancelFunc
}

func New[Bindings any](bindings *Bindings, opts ...Option[Bindings]) interfaces.ITakibi[Bindings] {
	if bindings == nil {
		bindings = new(Bindings)
	}

	ctx, cancel := stdContext.WithCancel(stdContext.Background())

	t := &takibi[Bindings]{
		env:    bindings,
		router: router.New[Bindings](),
		errorHandler: func(ctx interfaces.IContext[Bindings], err error) error {
			return ctx.Status(http.StatusInternalServerError).Text(err.Error())
		},
		ctx:          ctx,
		cancel:       cancel,
		maxBodyBytes: thttp.DefaultMaxBodyBytes,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (
	t *takibi[Bindings],
) Fire(
	addr string,
) error {
	workers.Serve(t)
	return nil
}

func (
	t *takibi[Bindings],
) startTasks() {
}

func (
	t *takibi[Bindings],
) Finish(
	ctx stdContext.Context,
) error {
	return errors.New("not support on wasm")
}

func (
	t *takibi[Bindings],
) stopTasks(ctx stdContext.Context) {
}

func (
	t *takibi[Bindings],
) setupServer() error {
	return nil
}

func (
	t *takibi[Bindings],
) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	// get from cache & reset context
	ctx := t.initializeContext(w, r)
	defer t.cache.Put(ctx)

	n, middlewares, params := t.router.Find(r.Method, r.URL.Path)
	if len(params) > 0 {
		ctx.SetParam(params)
	}

	var handler interfaces.HandlerFunc[Bindings]
	if n != nil {
		handler = n.ComposedHandler()
	}

	if handler == nil {
		notFound := func(c interfaces.IContext[Bindings]) error {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		handler = router.Compose(notFound, middlewares)
	}

	if err := handler(ctx); err != nil {
		if err := t.errorHandler(ctx, err); err != nil {
			// fallback
			ctx.Response().WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

func (
	t *takibi[Bindings],
) initializeContext(
	w http.ResponseWriter,
	r *http.Request,
) interfaces.IContext[Bindings] {
	if ctx, ok := t.cache.Get().(interfaces.IContext[Bindings]); ok {
		ctx.Reset(w, r)
		return ctx
	}

	return NewContext(w, r, t.Env(), t.maxBodyBytes)
}

func (
	t *takibi[Bindings],
) Env() *Bindings {
	return t.env
}

func (
	t *takibi[Bindings],
) OnError(
	handler interfaces.ErrorHandlerFunc[Bindings],
) {
	t.errorHandler = handler
}

func (
	t *takibi[Bindings],
) OnBlowError(
	handler interfaces.BlowErrorHandlerFunc[Bindings],
) {
	fmt.Println("it is not supported for wasm")
}

func (
	t *takibi[Bindings],
) Use(
	path string,
	middleware ...interfaces.MiddlewareFunc[Bindings],
) error {
	return t.router.Use(path, middleware...)
}

func (
	t *takibi[Bindings],
) Router() interfaces.IRouter[Bindings] {
	return t.router
}

func (
	t *takibi[Bindings],
) Route(
	basePath string,
	app interfaces.ITakibi[Bindings],
) error {
	linearRoutes := app.Router().LinearizeTree()
	for _, method := range router.SupportedHttpMethod {
		for _, route := range linearRoutes[method] {
			fullPath := basePath + route.Path

			if err := t.router.Add(method, fullPath, route.Handler); err != nil {
				return err
			}

			if err := t.router.Use(fullPath, route.Middleware...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (
	t *takibi[Bindings],
) Get(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Get(path, handler)
}

func (
	t *takibi[Bindings],
) Post(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Post(path, handler)
}

func (
	t *takibi[Bindings],
) Put(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Put(path, handler)
}

func (
	t *takibi[Bindings],
) Patch(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Patch(path, handler)
}

func (
	t *takibi[Bindings],
) Delete(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Delete(path, handler)
}

func (
	t *takibi[Bindings],
) Head(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Head(path, handler)
}

func (
	t *takibi[Bindings],
) Options(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Options(path, handler)
}

func (
	t *takibi[Bindings],
) Trace(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Trace(path, handler)
}

func (
	t *takibi[Bindings],
) Connect(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return t.router.Connect(path, handler)
}

func (
	t *takibi[Bindings],
) All(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	if err := t.router.Get(path, handler); err != nil {
		return err
	}
	if err := t.router.Post(path, handler); err != nil {
		return err
	}
	if err := t.router.Put(path, handler); err != nil {
		return err
	}
	if err := t.router.Patch(path, handler); err != nil {
		return err
	}
	if err := t.router.Delete(path, handler); err != nil {
		return err
	}
	if err := t.router.Head(path, handler); err != nil {
		return err
	}
	if err := t.router.Options(path, handler); err != nil {
		return err
	}
	if err := t.router.Trace(path, handler); err != nil {
		return err
	}
	if err := t.router.Connect(path, handler); err != nil {
		return err
	}
	return nil
}

func (
	t *takibi[Bindings],
) On(
	methods,
	paths []string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	for _, method := range methods {
		for _, path := range paths {
			switch method {
			case http.MethodGet:
				if err := t.Get(path, handler); err != nil {
					return err
				}
			case http.MethodPost:
				if err := t.Post(path, handler); err != nil {
					return err
				}
			case http.MethodPut:
				if err := t.Put(path, handler); err != nil {
					return err
				}
			case http.MethodPatch:
				if err := t.Patch(path, handler); err != nil {
					return err
				}
			case http.MethodDelete:
				if err := t.Delete(path, handler); err != nil {
					return err
				}
			case http.MethodHead:
				if err := t.Head(path, handler); err != nil {
					return err
				}
			case http.MethodOptions:
				if err := t.Options(path, handler); err != nil {
					return err
				}
			case http.MethodTrace:
				if err := t.Trace(path, handler); err != nil {
					return err
				}
			case http.MethodConnect:
				if err := t.Connect(path, handler); err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid method: %s", method)
			}
		}
	}
	return nil
}

func (
	t *takibi[Bindings],
) Blow(
	tasks ...interfaces.BlowTask[Bindings],
) {
	fmt.Println("it is not supported for wasm")
}

func (
	t *takibi[Bindings],
) Camp(
	method,
	path string,
	opts ...interfaces.CampOption,
) interfaces.ICampResponse {
	return nil
}
