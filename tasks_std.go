//go:build !wasm

package takibi

import (
	stdContext "context"
	"net/http"
	"sync"

	"github.com/poteto0/takibi/interfaces"
	"github.com/robfig/cron/v3"
)

func (t *takibi[Bindings]) startTasks() {
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

func (t *takibi[Bindings]) stopTasks(ctx stdContext.Context) {
	if t.cron != nil {
		t.cron.Stop()
	}

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
