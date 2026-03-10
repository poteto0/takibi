package factory

import (
	"fmt"
	"strconv"

	"github.com/poteto0/takibi/interfaces"
)

// CreateMiddleware creates a MiddlewareFunc from a function that takes a context and next handler, and returns a handler.
func CreateMiddleware[Bindings any](
	f func(interfaces.IContext[Bindings], interfaces.HandlerFunc[Bindings]) interfaces.HandlerFunc[Bindings],
) interfaces.MiddlewareFunc[Bindings] {
	return func(c interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) error {
		return f(c, next)(c)
	}
}

type ParamGetter interface {
	ParamBy(key string) string
}

func ParamBy[T any](c ParamGetter, key string) (T, error) {
	val := c.ParamBy(key)
	return convert[T](val)
}

type RequestGetter interface {
	Req() interfaces.IRequest
}

func QueryBy[T any](c RequestGetter, key string) (T, error) {
	val := c.Req().QueryBy(key)
	return convert[T](val)
}

func convert[T any](s string) (T, error) {
	var result T
	if s == "" {
		return result, nil // Return zero value if empty? Or error? Usually zero value is safer but maybe ambiguous.
	}

	switch any(result).(type) {
	case int:
		i, err := strconv.Atoi(s)
		if err != nil {
			return result, err
		}
		return any(i).(T), nil
	case int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return result, err
		}
		return any(i).(T), nil
	case float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return result, err
		}
		return any(f).(T), nil
	case bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return result, err
		}
		return any(b).(T), nil
	case string:
		return any(s).(T), nil
	default:
		return result, fmt.Errorf("unsupported type")
	}
}
