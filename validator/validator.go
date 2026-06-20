package validator

import (
	"net/url"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

// ErrStop aliases constants.ErrStop for convenience — return it from a
// validator fn to halt the handler chain without triggering the error handler.
// The response must be written by the returning handler before returning ErrStop.
var ErrStop = constants.ErrStop

// Target keys used by the built-in validator factories.
const (
	TargetForm  = "form"
	TargetJson  = "json"
	TargetQuery = "query"
	TargetParam = "param"
)

func newValidator[Bindings any, In any, T any](
	key string,
	extract func(interfaces.IContext[Bindings]) (In, error),
	fn func(In, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return func(c interfaces.IContext[Bindings]) error {
		input, err := extract(c)
		if err != nil {
			return err
		}
		result, err := fn(input, c)
		if err != nil {
			return err
		}
		c.SetValidated(key, result)
		return nil
	}
}

// Form returns a HandlerFunc that parses the request's form body and passes
// the values to fn. The returned value is stored under TargetForm ("form").
func Form[Bindings any, T any](
	fn func(url.Values, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetForm, func(c interfaces.IContext[Bindings]) (url.Values, error) {
		if err := c.Req().Raw().ParseForm(); err != nil {
			return nil, err
		}
		return c.Req().Raw().Form, nil
	}, fn)
}

// Json returns a HandlerFunc that reads the JSON request body and passes
// the parsed map to fn. The returned value is stored under TargetJson ("json").
func Json[Bindings any, T any](
	fn func(map[string]any, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetJson, func(c interfaces.IContext[Bindings]) (map[string]any, error) {
		return c.Req().Json()
	}, fn)
}

// Query returns a HandlerFunc that passes the request's query parameters to
// fn. The returned value is stored under TargetQuery ("query").
func Query[Bindings any, T any](
	fn func(map[string]string, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetQuery, func(c interfaces.IContext[Bindings]) (map[string]string, error) {
		return c.Req().Query(), nil
	}, fn)
}

// Unmarshall returns a HandlerFunc that decodes the JSON request body into a
// value of type T and passes it to fn. This is the typical typed-body pattern:
// fn receives a fully populated T rather than a map. The returned value is
// stored under TargetJson ("json").
func Unmarshall[Bindings any, T any](
	fn func(T, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetJson, func(c interfaces.IContext[Bindings]) (T, error) {
		var dest T
		if err := c.Req().Unmarshall(&dest); err != nil {
			return dest, err
		}
		return dest, nil
	}, fn)
}

// Param returns a HandlerFunc that passes the request's path parameters to
// fn. The returned value is stored under TargetParam ("param").
func Param[Bindings any, T any](
	fn func(map[string]string, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetParam, func(c interfaces.IContext[Bindings]) (map[string]string, error) {
		return c.Param(), nil
	}, fn)
}

// Valid retrieves the validated value stored under target and type-asserts it
// to T. Returns the zero value and false when the key is absent or the type
// does not match.
func Valid[T any](c interface {
	Validated(string) (any, bool)
}, target string) (T, bool) {
	v, ok := c.Validated(target)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}
