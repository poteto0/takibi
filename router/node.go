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

func (n *node[Bindings]) rebuildComposedHandlers() {
	n.walkAndCompose(nil)
}

// walkAndCompose is the recursive DFS worker for rebuildComposedHandlers.
// accumulated holds middlewares collected from ancestor nodes.
func (n *node[Bindings]) walkAndCompose(accumulated []interfaces.MiddlewareFunc[Bindings]) {
	// Three-index slice caps capacity so append always allocates a fresh backing
	// array, preventing sibling branches from sharing and corrupting each other's slice.
	full := append(accumulated[:len(accumulated):len(accumulated)], n.middlewares...)
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
		n.rebuildComposedHandlers()
		return nil
	}

	rightPath := path[1:]
	param := ""

	for {
		id := strings.Index(rightPath, "/")
		if id < 0 {
			param = rightPath
		} else {
			param = rightPath[:id]
			rightPath = rightPath[(id + 1):]
		}

		if child := currentNode.children[param]; child == nil {
			if hasPathParamPrefix(param) {
				currentNode.childParamKey = param
			}

			currentNode.children[param] = NewNode[Bindings]().(*node[Bindings])
		}

		currentNode = currentNode.children[param].(*node[Bindings])

		if id < 0 {
			break
		}
	}

	currentNode.middlewares = append(currentNode.middlewares, middleware...)
	n.rebuildComposedHandlers()
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
		n.rebuildComposedHandlers()
		return nil
	}

	param := ""

	for {
		id := strings.Index(rightPath, "/")
		if id < 0 {
			param = rightPath
		} else {
			param = rightPath[:id]
			rightPath = rightPath[(id + 1):]
		}

		if child := currentNode.children[param]; child == nil {
			if hasPathParamPrefix(param) {
				currentNode.childParamKey = param
			}

			currentNode.children[param] = NewNode[Bindings]().(*node[Bindings])
		}

		currentNode = currentNode.children[param].(*node[Bindings])

		if id < 0 {
			break
		}
	}

	if currentNode.handler != nil {
		return constants.ErrHandlerAlreadyExists
	}

	currentNode.handler = handler
	n.rebuildComposedHandlers()
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
	param := ""
	pathParams := map[string]string{}
	var middlewares []interfaces.MiddlewareFunc[Bindings]

	// Collect root middlewares
	middlewares = append(middlewares, currentNode.middlewares...)

	if rightPath == "" {
		return currentNode, middlewares, pathParams
	}

	for {
		id := strings.Index(rightPath, "/")
		if id < 0 {
			param = rightPath
		} else {
			param = rightPath[:id]
			rightPath = rightPath[(id + 1):]
		}

		if child := currentNode.children[param]; child != nil {
			currentNode = child.(*node[Bindings])
			middlewares = append(middlewares, currentNode.middlewares...)
			if id < 0 {
				break
			}
			continue
		}

		// param key
		if chParam := currentNode.childParamKey; chParam != "" {
			if chNode := currentNode.children[chParam]; chNode != nil {
				currentNode = chNode.(*node[Bindings])
				middlewares = append(middlewares, currentNode.middlewares...)
				pathParams[chParam[1:]] = param
			}
		} else {
			// not found
			return nil, middlewares, pathParams
		}

		if id < 0 {
			break
		}
	}

	return currentNode, middlewares, pathParams
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
