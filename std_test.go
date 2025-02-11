//go:build goexperiment.arenas

package arena

import (
	sa "arena"
	"runtime"
	"runtime/debug"
	"testing"
)

func BenchmarkStd_Arena_Alloc_int(b *testing.B) {
	a := sa.NewArena()
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sa.New[int](a)
	}
	a.Free()
	runtime.GC()
	debug.FreeOSMemory()
}

func BenchmarkStd_Arena_Alloc_intAndFree(b *testing.B) {
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := sa.NewArena()
		for i := 0; i < 100; i++ {
			sa.New[int](a)
		}
		a.Free()
	}
}
