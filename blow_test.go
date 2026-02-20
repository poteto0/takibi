package takibi_test

import (
	"context"
	"testing"
	"time"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestBlow_TriggerStart(t *testing.T) {
	app := takibi.New(&struct{}{})
	executed := false
	app.Blow(interfaces.BlowTask[struct{}]{
		BlowActionTag:     "trigger",
		BlowActionTrigger: "start",
		BlowAction: func(ctx context.Context, env *struct{}) error {
			executed = true
			return nil
		},
	})

	// Start server in background
	go func() {
		_ = app.Fire(":0")
	}()

	// Wait for execution
	assert.Eventually(t, func() bool {
		return executed
	}, 1*time.Second, 10*time.Millisecond)

	_ = app.Finish(context.Background())
}

func TestBlow_TriggerStop(t *testing.T) {
	app := takibi.New(&struct{}{})
	executed := false
	app.Blow(interfaces.BlowTask[struct{}]{
		BlowActionTag:     "trigger",
		BlowActionTrigger: "stop",
		BlowAction: func(ctx context.Context, env *struct{}) error {
			executed = true
			return nil
		},
	})

	// Start server in background
	go func() {
		_ = app.Fire(":0")
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	_ = app.Finish(context.Background())

	// Wait for execution
	assert.Eventually(t, func() bool {
		return executed
	}, 1*time.Second, 10*time.Millisecond)
}

func TestBlow_Schedule(t *testing.T) {
	app := takibi.New(&struct{}{})
	executed := false
	app.Blow(interfaces.BlowTask[struct{}]{
		BlowActionTag:      "schedule",
		BlowActionSchedule: "@every 1s",
		BlowAction: func(ctx context.Context, env *struct{}) error {
			executed = true
			return nil
		},
	})

	// Start server in background
	go func() {
		_ = app.Fire(":0")
	}()

	// Wait for execution (cron @every 1s takes 1s to first run usually)
	assert.Eventually(t, func() bool {
		return executed
	}, 2*time.Second, 100*time.Millisecond)

	_ = app.Finish(context.Background())
}
