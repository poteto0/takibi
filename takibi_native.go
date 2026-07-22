//go:build !wasm

package takibi

import (
	stdContext "context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/router"
	"github.com/robfig/cron/v3"
)

type takibi[Bindings any] struct {
	env              *Bindings
	cache            sync.Pool
	router           interfaces.IRouter[Bindings]
	errorHandler     interfaces.ErrorHandlerFunc[Bindings]
	blowErrorHandler interfaces.BlowErrorHandlerFunc[Bindings]
	envResolver      interfaces.EnvResolverFunc[Bindings]
	tasks            []interfaces.BlowTask[Bindings]
	cron             *cron.Cron
	option           interfaces.TakibiOption

	ctx       stdContext.Context
	cancel    stdContext.CancelFunc
	fireMutex sync.RWMutex
	Server    http.Server
	Listener  net.Listener
}

func New[Bindings any](bindings *Bindings) interfaces.ITakibi[Bindings] {
	return NewWithOption(bindings, interfaces.DefaultTakibiOption)
}

func NewWithOption[Bindings any](bindings *Bindings, opt interfaces.TakibiOption) interfaces.ITakibi[Bindings] {
	if bindings == nil {
		bindings = new(Bindings)
	}

	ctx, cancel := stdContext.WithCancel(stdContext.Background())

	return &takibi[Bindings]{
		env:         bindings,
		envResolver: defaultEnvResolver(bindings),
		router:      router.New[Bindings](),
		errorHandler: func(ctx interfaces.IContext[Bindings], err error) error {
			return ctx.Status(http.StatusInternalServerError).Text("Internal Server Error")
		},
		blowErrorHandler: func(c interfaces.IContext[Bindings], err error) {
			fmt.Println(err.Error())
		},
		ctx:    ctx,
		cancel: cancel,
		option: opt,
	}
}

func (
	t *takibi[Bindings],
) Fire(
	addr string,
) error {
	t.fireMutex.Lock()

	if !strings.HasPrefix(addr, constants.PortPrefix) {
		addr = constants.PortPrefix + addr
	}

	t.Server.Addr = addr
	if err := t.setupServer(); err != nil {
		t.fireMutex.Unlock()
		return err
	}

	t.fireMutex.Unlock()

	t.startTasks()

	return t.Server.Serve(t.Listener)
}

func (
	t *takibi[Bindings],
) startTasks() {
	for _, task := range t.tasks {
		if task.BlowActionTag == interfaces.BlowTagTrigger && task.BlowActionTrigger == interfaces.BlowTriggerStart {
			r, _ := http.NewRequestWithContext(t.ctx, "GET", "/", nil)
			c := t.newTaskContext(r)
			go func(task interfaces.BlowTask[Bindings]) {
				if err := task.BlowAction(c); err != nil {
					t.blowErrorHandler(c, err)
				}
			}(task)
		}

		if task.BlowActionTag == interfaces.BlowTagSchedule && task.BlowActionSchedule != "" {
			if t.cron == nil {
				t.cron = cron.New(cron.WithSeconds())
			}
			_, _ = t.cron.AddFunc(task.BlowActionSchedule, func() {
				r, _ := http.NewRequestWithContext(t.ctx, "GET", "/", nil)
				c := t.newTaskContext(r)
				if err := task.BlowAction(c); err != nil {
					t.blowErrorHandler(c, err)
				}
			})
		}
	}

	if t.cron != nil {
		t.cron.Start()
	}
}

func (
	t *takibi[Bindings],
) Finish(
	ctx stdContext.Context,
) error {
	t.fireMutex.Lock()

	if err := t.Server.Shutdown(ctx); err != nil {
		t.fireMutex.Unlock()
		return err
	}

	t.stopTasks(ctx)

	t.cancel()

	t.fireMutex.Unlock()
	return nil
}

func (
	t *takibi[Bindings],
) stopTasks(ctx stdContext.Context) {
	if t.cron != nil {
		t.cron.Stop()
	}

	// execute stop tasks
	var wg sync.WaitGroup
	for _, task := range t.tasks {
		if task.BlowActionTag == interfaces.BlowTagTrigger && task.BlowActionTrigger == interfaces.BlowTriggerStop {
			wg.Add(1)
			go func(task interfaces.BlowTask[Bindings]) {
				defer wg.Done()
				r, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
				c := t.newTaskContext(r)
				if err := task.BlowAction(c); err != nil {
					t.blowErrorHandler(c, err)
				}
			}(task)
		}
	}
	wg.Wait()
}

func (
	t *takibi[Bindings],
) setupServer() error {
	// TODO: print banner

	// setting handler
	t.Server.Handler = t

	if t.Listener != nil {
		return nil
	}

	// set listener
	// TODO: add multiple listner
	ln, err := net.Listen("tcp", t.Server.Addr)
	if err != nil {
		return err
	}

	if t.Server.TLSConfig == nil {
		t.Listener = ln
		return nil
	}

	// tls mode
	t.Listener = tls.NewListener(ln, t.Server.TLSConfig)
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
	t.blowErrorHandler = handler
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
) Blow(
	tasks ...interfaces.BlowTask[Bindings],
) {
	t.tasks = append(t.tasks, tasks...)
}

func (
	t *takibi[Bindings],
) Camp(
	method,
	path string,
	opts ...interfaces.CampOption,
) interfaces.ICampResponse {
	r, _ := http.NewRequest(method, path, nil)
	for _, opt := range opts {
		opt(r)
	}

	w := httptest.NewRecorder()
	t.ServeHTTP(w, r)

	return newCampResponse(w.Result())
}
