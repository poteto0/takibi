package interfaces

import (
	"context"
)

type BlowTask[Bindings any] struct {
	BlowActionTag      string // "trigger" or "schedule"
	BlowActionSchedule string // cron schedule or empty
	BlowActionTrigger  string // "start" or "stop"
	BlowAction         func(ctx context.Context, env *Bindings) error
}
