package middlewares

import (
	"fmt"
	"time"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

func Timeout[Bindings any](limit time.Duration) interfaces.MiddlewareFunc[Bindings] {
	return func(ctx interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) error {
		var result error

		done := make(chan struct{})
		go func() {
			defer func() {
				// in case of panic
				if r := recover(); r != nil {
					result = fmt.Errorf("%v", r)
				}

				close(done)
			}()

			// do
			result = next(ctx)
		}()

		select {
		case <-done:
			return result
		// this loaded
		case <-time.After(limit):
			return constants.ErrRequestTimeout
		}
	}
}
