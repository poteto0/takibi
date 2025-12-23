package router

import "net/http"

var SupportedHttpMethod = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodHead,
	http.MethodOptions,
	http.MethodConnect,
	http.MethodTrace,
}
