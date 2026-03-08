package interfaces

import "net/http"

type IRequest interface {
	Raw() *http.Request

	/* Headers */
	Header() http.Header
	HeaderBy(key string) string
	ContentType() string

	/* Parameters */
	// get request body as map
	Json() (map[string]any, error)

	// Unmarshall request body to dest
	Unmarshall(dest any) error

	// get query parameters as map
	Queries() map[string][]string

	// get query parameters by key
	QueriesBy(key string) []string

	// get query parameters as map
	// if multiple values exist, only the first one is returned
	Query() map[string]string

	// get query parameter by key
	// if multiple values exist, only the first one is returned
	QueryBy(key string) string
}
