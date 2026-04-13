//go:build !wasm

package takibi

import (
	"net/http"
	"net/http/httptest"

	"github.com/poteto0/takibi/interfaces"
)

func (t *takibi[Bindings]) Camp(method, path string, opts ...interfaces.CampOption) interfaces.ICampResponse {
	r, _ := http.NewRequest(method, path, nil)
	for _, opt := range opts {
		opt(r)
	}

	w := httptest.NewRecorder()
	t.ServeHTTP(w, r)

	return newCampResponse(w.Result())
}
