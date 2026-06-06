package router

import (
	"strconv"
	"testing"

	"github.com/poteto0/takibi/interfaces"
)

func benchRegNopHandler(ctx interfaces.IContext[any]) error { return nil }

// benchmarkRegister registers numRoutes distinct routes from scratch, exercising
// the per-registration handler (re)composition. The old implementation
// re-composed the entire tree on every Add (O(N^2)); the new one recomposes
// only the inserted subtree (O(N) overall).
func benchmarkRegister(b *testing.B, numRoutes int) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr := New[any]()
		for j := 0; j < numRoutes; j++ {
			_ = tr.Get("/route/"+strconv.Itoa(j), benchRegNopHandler)
		}
	}
}

func BenchmarkRegister_100(b *testing.B)  { benchmarkRegister(b, 100) }
func BenchmarkRegister_500(b *testing.B)  { benchmarkRegister(b, 500) }
func BenchmarkRegister_1000(b *testing.B) { benchmarkRegister(b, 1000) }
func BenchmarkRegister_2000(b *testing.B) { benchmarkRegister(b, 2000) }
