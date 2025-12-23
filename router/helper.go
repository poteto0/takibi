package router

import (
	"slices"
	"strings"

	"github.com/poteto0/takibi/constants"
)

func hasPathParamPrefix(path string) bool {
	if len(path) == 0 {
		return false
	}

	return strings.HasPrefix(
		path,
		constants.PathPramPrefix,
	)
}

func isSupportedHttpMethod(method string) bool {
	if len(method) == 0 {
		return false
	}

	if slices.Contains(
		SupportedHttpMethod,
		method,
	) {
		return true
	}

	return false
}
