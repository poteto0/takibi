package router

import (
	"strings"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

type node[Bindings any] struct {
	children        map[string]interfaces.INode[Bindings]
	childParamKey   string
	handler         interfaces.HandlerFunc[Bindings]
	composedHandler interfaces.HandlerFunc[Bindings]
	middlewares     []interfaces.MiddlewareFunc[Bindings]
}

func NewNode[Bindings any]() interfaces.INode[Bindings] {
	return &node[Bindings]{
		children: make(map[string]interfaces.INode[Bindings]),
	}
}

func (
	n *node[Bindings],
) Handler() interfaces.HandlerFunc[Bindings] {
	return n.handler
}

func (
	n *node[Bindings],
) ComposedHandler() interfaces.HandlerFunc[Bindings] {
	return n.composedHandler
}

// Compose wraps handler with middlewares, outermost first.
func Compose[Bindings any](
	handler interfaces.HandlerFunc[Bindings],
	middlewares []interfaces.MiddlewareFunc[Bindings],
) interfaces.HandlerFunc[Bindings] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := handler
		handler = func(ctx interfaces.IContext[Bindings]) error {
			return mw(ctx, next)
		}
	}
	return handler
}

// walkAndCompose recomposes the subtree rooted at n.
// accumulated holds middlewares collected from ancestor nodes (excluding n).
// Callers pass the ancestor middlewares so only the affected subtree is
// rebuilt, instead of re-composing the whole tree on every registration.
func (n *node[Bindings]) walkAndCompose(accumulated []interfaces.MiddlewareFunc[Bindings]) {
	full := make([]interfaces.MiddlewareFunc[Bindings], len(accumulated)+len(n.middlewares))
	copy(full, accumulated)
	copy(full[len(accumulated):], n.middlewares)
	if n.handler != nil {
		n.composedHandler = Compose(n.handler, full)
	}
	for _, child := range n.children {
		child.(*node[Bindings]).walkAndCompose(full)
	}
}

func (
	n *node[Bindings],
) Middlewares() []interfaces.MiddlewareFunc[Bindings] {
	return n.middlewares
}

func (
	n *node[Bindings],
) AddMiddleware(
	path string,
	middleware ...interfaces.MiddlewareFunc[Bindings],
) error {
	currentNode := n
	// handle "*" or "/*" suffix
	if path == "*" {
		path = "/"
	}
	path = strings.TrimSuffix(path, "/*")

	// ensure path starts with / if not empty
	if path != "" && path[0] != '/' {
		return constants.ErrInvalidPath
	}

	// if path is just / or empty after trim, it's root
	if path == "/" || path == "" {
		currentNode.middlewares = append(currentNode.middlewares, middleware...)
		currentNode.walkAndCompose(nil)
		return nil
	}

	rightPath := path[1:]
	var accumulated []interfaces.MiddlewareFunc[Bindings]

	for {
		param, rest := nextSegment(rightPath)

		if child := currentNode.children[param]; child == nil {
			if hasPathParamPrefix(param) {
				currentNode.childParamKey = param
			}

			currentNode.children[param] = NewNode[Bindings]().(*node[Bindings])
		}

		accumulated = append(accumulated, currentNode.middlewares...)
		currentNode = currentNode.children[param].(*node[Bindings])

		if rest == "" {
			break
		}
		rightPath = rest
	}

	currentNode.middlewares = append(currentNode.middlewares, middleware...)
	currentNode.walkAndCompose(accumulated)
	return nil
}

func (
	n *node[Bindings],
) Add(
	path string,
	handler interfaces.HandlerFunc[Bindings],
) error {
	currentNode := n
	rightPath := path[1:]

	if rightPath == "" {
		if currentNode.handler != nil {
			return constants.ErrHandlerAlreadyExists
		}
		currentNode.handler = handler
		currentNode.walkAndCompose(nil)
		return nil
	}

	var accumulated []interfaces.MiddlewareFunc[Bindings]

	for {
		param, rest := nextSegment(rightPath)

		if child := currentNode.children[param]; child == nil {
			if hasPathParamPrefix(param) {
				currentNode.childParamKey = param
			}

			currentNode.children[param] = NewNode[Bindings]().(*node[Bindings])
		}

		accumulated = append(accumulated, currentNode.middlewares...)
		currentNode = currentNode.children[param].(*node[Bindings])

		if rest == "" {
			break
		}
		rightPath = rest
	}

	if currentNode.handler != nil {
		return constants.ErrHandlerAlreadyExists
	}

	currentNode.handler = handler
	currentNode.walkAndCompose(accumulated)
	return nil
}

func (
	n *node[Bindings],
) Find(
	path string,
) (
	interfaces.INode[Bindings],
	[]interfaces.MiddlewareFunc[Bindings],
	map[string]string,
) {
	currentNode := n
	rightPath := path[1:]
	// pathParams and middlewares are allocated lazily: a matched static route
	// (the hot path) uses the node's pre-composed handler and needs neither.
	var pathParams map[string]string

	for rightPath != "" {
		var param string
		param, rightPath = nextSegment(rightPath)

		if child := currentNode.children[param]; child != nil {
			currentNode = child.(*node[Bindings])
			continue
		}

		// param key
		if chParam := currentNode.childParamKey; chParam != "" {
			if chNode := currentNode.children[chParam]; chNode != nil {
				currentNode = chNode.(*node[Bindings])
				if pathParams == nil {
					pathParams = map[string]string{}
				}
				pathParams[chParam[1:]] = param
				continue
			}
		}

		// not found: collect prefix middlewares for the caller's 404 handler.
		return nil, n.collectMiddlewares(path), pathParams
	}

	if currentNode.handler == nil {
		// Matched an intermediate node without a handler. The caller treats
		// this like not-found, so it still needs the prefix middlewares.
		return currentNode, n.collectMiddlewares(path), pathParams
	}

	return currentNode, nil, pathParams
}

// collectMiddlewares walks path and accumulates the middlewares along the
// matched prefix, outermost (root) first. It is only used on the cold
// not-found / handler-less paths, so its allocations never hit matched routes.
func (n *node[Bindings]) collectMiddlewares(
	path string,
) []interfaces.MiddlewareFunc[Bindings] {
	currentNode := n
	middlewares := append(
		[]interfaces.MiddlewareFunc[Bindings]{},
		currentNode.middlewares...,
	)
	rightPath := path[1:]

	for rightPath != "" {
		var param string
		param, rightPath = nextSegment(rightPath)

		next := currentNode.children[param]
		if next == nil {
			if chParam := currentNode.childParamKey; chParam != "" {
				next = currentNode.children[chParam]
			}
		}
		if next == nil {
			break
		}

		currentNode = next.(*node[Bindings])
		middlewares = append(middlewares, currentNode.middlewares...)
	}

	return middlewares
}

func (
	n *node[Bindings],
) Linearize() []interfaces.NodeUnit[Bindings] {
	visited := make(map[string]struct{})
	results := []interfaces.NodeUnit[Bindings]{}
	n.dfs(n, "", &visited, &results)
	return results
}

func (
	n *node[Bindings],
) dfs(
	curr *node[Bindings],
	path string,
	visited *map[string]struct{},
	results *[]interfaces.NodeUnit[Bindings],
) {
	if curr == nil {
		return
	}

	if _, ok := (*visited)[path]; ok {
		return
	}
	(*visited)[path] = struct{}{}

	if curr.handler != nil {
		*results = append(*results, interfaces.NodeUnit[Bindings]{
			Path:       path,
			Handler:    curr.handler,
			Middleware: curr.middlewares,
		})
	}

	for key, child := range curr.children {
		childNode := child.(*node[Bindings])
		nextPath := path + "/" + key
		n.dfs(childNode, nextPath, visited, results)
	}
}
