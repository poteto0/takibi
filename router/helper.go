package router

import (
	"slices"
	"strings"

	"github.com/poteto0/takibi/constants"
)

// nextSegment splits the leading path segment off rightPath. It returns the
// segment and the remaining path, which is "" once the last segment is
// reached. Callers loop while the remaining path is non-empty.
func nextSegment(rightPath string) (segment, rest string) {
	if id := strings.Index(rightPath, "/"); id >= 0 {
		return rightPath[:id], rightPath[id+1:]
	}
	return rightPath, ""
}

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
