package takibi

import (
	stdContext "context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/router"
	"github.com/robfig/cron/v3"
)

type takibi[Bindings any] struct {
	env          *Bindings
	cache        sync.Pool
		router           interfaces.IRouter[Bindings]
		errorHandler     interfaces.ErrorHandlerFunc[Bindings]
		blowErrorHandler interfaces.BlowErrorHandlerFunc[Bindings]
		tasks            []interfaces.BlowTask[Bindings]
		cron             *cron.Cron
	
	ctx          stdContext.Context
	cancel       stdContext.CancelFunc
	fireMutex    sync.RWMutex
	Server       http.Server
	Listener     net.Listener
}

func New[Bindings any](bindings *Bindings) interfaces.ITakibi[Bindings] {
	if bindings == nil {
		bindings = new(Bindings)
	}

	ctx, cancel := stdContext.WithCancel(stdContext.Background())

	return &takibi[Bindings]{
		env:    bindings,
		router: router.New[Bindings](),
		errorHandler: func(ctx interfaces.IContext[Bindings], err error) error {
			return ctx.Status(http.StatusInternalServerError).Text(err.Error())
		},
		blowErrorHandler: func(c interfaces.IContext[Bindings], err error) {
			fmt.Println(err.Error())
		},
		ctx:    ctx,
		cancel: cancel,
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
		if task.BlowActionTag == "trigger" && task.BlowActionTrigger == "start" {
			r, _ := http.NewRequestWithContext(t.ctx, "GET", "/", nil)
			c := NewContext(nil, r, t.env)
			go func(task interfaces.BlowTask[Bindings]) {
				if err := task.BlowAction(c); err != nil {
					t.blowErrorHandler(c, err)
				}
			}(task)
		}

		if task.BlowActionTag == "schedule" && task.BlowActionSchedule != "" {
			if t.cron == nil {
				t.cron = cron.New(cron.WithSeconds())
			}
			_, _ = t.cron.AddFunc(task.BlowActionSchedule, func() {
				r, _ := http.NewRequestWithContext(t.ctx, "GET", "/", nil)
				c := NewContext(nil, r, t.env)
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
		if task.BlowActionTag == "trigger" && task.BlowActionTrigger == "stop" {
			wg.Add(1)
			go func(task interfaces.BlowTask[Bindings]) {
				defer wg.Done()
				r, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
				c := NewContext(nil, r, t.env)
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

	// TODO: pathParams
	n, middlewares, _ := t.router.Find(r.Method, r.URL.Path)

	var handler interfaces.HandlerFunc[Bindings]
	if n != nil {
		handler = n.Handler()
	}

	if handler == nil {
		handler = func(c interfaces.IContext[Bindings]) error {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
	}

	composedHandler := compose(handler, middlewares)

	if err := composedHandler(ctx); err != nil {
		if err := t.errorHandler(ctx, err); err != nil {
			// fallback
			ctx.Response().WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

func compose[Bindings any](
	handler interfaces.HandlerFunc[Bindings],
	middlewares []interfaces.MiddlewareFunc[Bindings],
) interfaces.HandlerFunc[Bindings] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := handler
		handler = func(ctx interfaces.IContext[Bindings]) error {
			return mw(ctx, next)
		}
	}
	return handler
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

	return NewContext(w, r, t.Env())
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
) Blow(
	task interfaces.BlowTask[Bindings],
) {
	t.tasks = append(t.tasks, task)
}
