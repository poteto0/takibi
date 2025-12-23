package router

import (
	"errors"
	"net/http"
	"strings"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

type trieRouter[Bindings any] struct {
	trees map[string]interfaces.INode[Bindings]
}

func New[Bindings any]() interfaces.IRouter[Bindings] {
	return &trieRouter[Bindings]{
		trees: map[string]interfaces.INode[Bindings]{
			http.MethodGet:     NewNode[Bindings](),
			http.MethodPost:    NewNode[Bindings](),
			http.MethodPut:     NewNode[Bindings](),
			http.MethodPatch:   NewNode[Bindings](),
			http.MethodDelete:  NewNode[Bindings](),
			http.MethodHead:    NewNode[Bindings](),
			http.MethodOptions: NewNode[Bindings](),
			http.MethodTrace:   NewNode[Bindings](),
			http.MethodConnect: NewNode[Bindings](),
		},
	}
}

func (
	tr *trieRouter[Bindings],
) Find(
	method,
	path string,
) (
	interfaces.INode[Bindings],
	[]interfaces.MiddlewareFunc[Bindings],
	map[string]string,
) {
	tree := tr.trees[method]
	return tree.Find(path)
}

func (
	tr *trieRouter[Bindings],
) Use(
	path string,
	middleware ...interfaces.MiddlewareFunc[Bindings],
) error {
	for _, tree := range tr.trees {
		if err := tree.AddMiddleware(path, middleware...); err != nil {
			return err
		}
	}
	return nil
}

func (
	tr *trieRouter[Bindings],
) add(
	method,
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	if !isSupportedHttpMethod(method) {
		return constants.ErrNotSupportedMethod
	}

	if path != "/" {
		// "/users/" -> "/users"
		// if just "/" -> handler set by above
		path = strings.TrimSuffix(path, "/")
	}

	err := tr.trees[method].Add(path, handler)
	if err != nil {
		return errors.Join(
			err,
			errors.New("["+method+"] "+path+" is already used"),
		)
	}

	return nil
}

func (
	tr *trieRouter[Bindings],
) Get(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodGet, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Post(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodPost, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Put(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodPut, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Patch(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodPatch, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Delete(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodDelete, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Head(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodHead, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Options(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodOptions, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Trace(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodTrace, path, handler)
}

func (
	tr *trieRouter[Bindings],
) Connect(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	return tr.add(http.MethodConnect, path, handler)
}
