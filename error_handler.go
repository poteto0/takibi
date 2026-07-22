package takibi

import (
	"github.com/poteto0/takibi/interfaces"
)

// handleError dispatches to the handler registered by OnError, falling back to
// the build-specific defaultErrorHandler. errorHandler stays nil until OnError
// sets one, which is also how Route tells whether a sub app customized it.
func (t *takibi[Bindings]) handleError(ctx interfaces.IContext[Bindings], err error) error {
	if t.errorHandler == nil {
		return defaultErrorHandler(ctx, err)
	}
	return t.errorHandler(ctx, err)
}

// wrapWithErrorHandler lets a sub app handle its own errors with the handler it
// registered via OnError. An error returned by that handler still flows to the
// parent's error handler.
func wrapWithErrorHandler[Bindings any](
	handler interfaces.HandlerFunc[Bindings],
	errorHandler interfaces.ErrorHandlerFunc[Bindings],
) interfaces.HandlerFunc[Bindings] {
	return func(c interfaces.IContext[Bindings]) error {
		if err := handler(c); err != nil {
			return errorHandler(c, err)
		}
		return nil
	}
}
