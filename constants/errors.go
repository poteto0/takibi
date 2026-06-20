package constants

import "errors"

var (
	ErrHandlerAlreadyExists = errors.New(
		"handler already exists",
	)

	ErrNotSupportedMethod = errors.New(
		"not supported method",
	)

	ErrInvalidPath = errors.New(
		"invalid path",
	)

	ErrParamMissing = errors.New(
		"parameter is missing",
	)

	ErrNoHandler = errors.New(
		"at least one handler is required",
	)
)

var (
	ErrRequestTimeout = errors.New(
		"request timeout",
	)
)

// ErrStop is a sentinel that a handler chain element returns to halt further
// handlers without triggering the framework error handler. The response must
// be written by the returning handler before returning ErrStop.
var ErrStop = errors.New("handler: stop")
