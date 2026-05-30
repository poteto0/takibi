package takibi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/takibi/interfaces"
)

func nopHandler(c interfaces.IContext[any]) error { return nil }

func nopMiddleware(c interfaces.IContext[any], next interfaces.HandlerFunc[any]) error {
	return next(c)
}

func newBenchApp(numMiddlewares int) interfaces.ITakibi[any] {
	app := New[any](nil)
	for i := 0; i < numMiddlewares; i++ {
		app.Use("*", nopMiddleware)
	}
	app.Get("/bench", nopHandler)
	return app
}

func BenchmarkServeHTTP_0MW(b *testing.B) {
	app := newBenchApp(0)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTP_1MW(b *testing.B) {
	app := newBenchApp(1)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTP_5MW(b *testing.B) {
	app := newBenchApp(5)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTP_10MW(b *testing.B) {
	app := newBenchApp(10)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func newResetBenchCtx(b *testing.B) (interfaces.IContext[any], *httptest.ResponseRecorder, *http.Request) {
	b.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewContext[any](httptest.NewRecorder(), req, nil)
	newReq := httptest.NewRequest(http.MethodGet, "/new", nil)
	return ctx, httptest.NewRecorder(), newReq
}

func BenchmarkContextReset(b *testing.B) {
	ctx, newW, newReq := newResetBenchCtx(b)
	ctx.SetParam(map[string]string{"id": "1", "name": "foo"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Reset(newW, newReq)
	}
}

func BenchmarkContextReset_WithParams(b *testing.B) {
	ctx, newW, newReq := newResetBenchCtx(b)
	params := map[string]string{"id": "1", "name": "foo"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.SetParam(params)
		ctx.Reset(newW, newReq)
	}
}
