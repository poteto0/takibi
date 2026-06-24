package interfaces

// BlowActionTag values.
const (
	BlowTagTrigger  = "trigger"
	BlowTagSchedule = "schedule"
)

// BlowActionTrigger values.
const (
	BlowTriggerStart = "start"
	BlowTriggerStop  = "stop"
)

type BlowTask[Bindings any] struct {
	BlowActionTag      string // BlowTagTrigger or BlowTagSchedule
	BlowActionSchedule string // cron schedule or empty
	BlowActionTrigger  string // BlowTriggerStart or BlowTriggerStop
	BlowAction         func(c IContext[Bindings]) error
}
