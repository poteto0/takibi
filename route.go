package takibi

import (
	"errors"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

// register chains the handlers and delegates to add. It returns
// constants.ErrNoHandler when no handler is provided.
func register[Bindings any](
	add func(string, interfaces.HandlerFunc[Bindings]) error,
	path string,
	handlers []interfaces.HandlerFunc[Bindings],
) error {
	if len(handlers) == 0 {
		return constants.ErrNoHandler
	}
	return add(path, chainHandlers(handlers))
}

func chainHandlers[Bindings any](handlers []interfaces.HandlerFunc[Bindings]) interfaces.HandlerFunc[Bindings] {
	nonNil := handlers[:0:0]
	for _, h := range handlers {
		if h != nil {
			nonNil = append(nonNil, h)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	return func(c interfaces.IContext[Bindings]) error {
		for _, h := range nonNil {
			if err := h(c); err != nil {
				if errors.Is(err, constants.ErrStop) {
					return nil
				}
				return err
			}
		}
		return nil
	}
}

func (
	t *takibi[Bindings],
) Get(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Get, path, handlers)
}

func (
	t *takibi[Bindings],
) Post(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Post, path, handlers)
}

func (
	t *takibi[Bindings],
) Put(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Put, path, handlers)
}

func (
	t *takibi[Bindings],
) Patch(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Patch, path, handlers)
}

func (
	t *takibi[Bindings],
) Delete(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Delete, path, handlers)
}

func (
	t *takibi[Bindings],
) Head(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Head, path, handlers)
}

func (
	t *takibi[Bindings],
) Options(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Options, path, handlers)
}

func (
	t *takibi[Bindings],
) Trace(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Trace, path, handlers)
}

func (
	t *takibi[Bindings],
) Connect(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	return register(t.router.Connect, path, handlers)
}

func (
	t *takibi[Bindings],
) All(
	path string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	if len(handlers) == 0 {
		return constants.ErrNoHandler
	}
	h := chainHandlers(handlers)
	for _, add := range []func(string, interfaces.HandlerFunc[Bindings]) error{
		t.router.Get, t.router.Post, t.router.Put, t.router.Patch,
		t.router.Delete, t.router.Head, t.router.Options, t.router.Trace, t.router.Connect,
	} {
		if err := add(path, h); err != nil {
			return err
		}
	}
	return nil
}

func (
	t *takibi[Bindings],
) On(
	methods,
	paths []string,
	handlers ...interfaces.HandlerFunc[Bindings],
) error {
	if len(handlers) == 0 {
		return constants.ErrNoHandler
	}
	h := chainHandlers(handlers)
	for _, method := range methods {
		for _, path := range paths {
			if err := t.router.Add(method, path, h); err != nil {
				return err
			}
		}
	}
	return nil
}
