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
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/cron"
)

type takibi[Bindings any] struct {
	env    *Bindings
	cache  sync.Pool
	router interfaces.IRouter[Bindings]
	// errorHandler is nil until OnError sets one, see handleError
	errorHandler     interfaces.ErrorHandlerFunc[Bindings]
	blowErrorHandler interfaces.BlowErrorHandlerFunc[Bindings]
	tasks            []interfaces.BlowTask[Bindings]
	option           interfaces.TakibiOption

	ctx    stdContext.Context
	cancel stdContext.CancelFunc
}

// defaultErrorHandler is used until OnError installs one.
func defaultErrorHandler[Bindings any](ctx interfaces.IContext[Bindings], err error) error {
	return ctx.Status(http.StatusInternalServerError).Text(err.Error())
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
		env:    bindings,
		router: router.New[Bindings](),
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
	workers.ServeNonBlock(t)
	t.startTasks()
	workers.Ready()
	<-workers.Done()
	return nil
}

// startTasks registers a single Cron Trigger dispatcher when at least one
// "schedule" task is registered. On Cloudflare Workers the actual firing
// schedule is defined by wrangler.jsonc `triggers.crons`; each fired event's
// cron expression is matched against BlowActionSchedule.
func (
	t *takibi[Bindings],
) startTasks() {
	var scheduleTasks []interfaces.BlowTask[Bindings]
	for _, task := range t.tasks {
		if task.BlowActionTag == interfaces.BlowTagSchedule {
			scheduleTasks = append(scheduleTasks, task)
		}
	}
	if len(scheduleTasks) == 0 {
		return
	}

	cron.ScheduleTaskNonBlock(func(ctx stdContext.Context) error {
		event, _ := cron.NewEvent(ctx)
		r, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
		c := NewContext(nil, r, t.env, &t.option)
		for _, task := range scheduleTasks {
			// When BlowActionSchedule is set, only run for the matching
			// fired cron expression. An empty schedule runs on every event.
			if task.BlowActionSchedule != "" && event != nil &&
				task.BlowActionSchedule != event.Cron {
				continue
			}
			if err := task.BlowAction(c); err != nil {
				t.blowErrorHandler(c, err)
			}
		}
		return nil
	})
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
		if err := t.handleError(ctx, err); err != nil {
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

	return NewContext(w, r, t.Env(), &t.option)
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
	return nil
}
