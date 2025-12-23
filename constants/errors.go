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
)

var (
	ErrRequestTimeout = errors.New(
		"request timeout",
	)
)
