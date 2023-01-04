package arena

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

const n = 100000

type user struct {
	a int
	b int
	c int
}

func TestAlloc(t *testing.T) {
	var value, check []*user = make([]*user, n), make([]*user, n)
	a := NewArena(unsafe.Sizeof(user{}))
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	var sema = make(chan struct{}, 8100)
	for i := 0; i < 8100; i++ {
		sema <- struct{}{}
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		<-sema
		go func(gi int) {
			defer wg.Done()
			u := (*user)(a.Alloc())
			u.a = 1
			u.b = 2
			u.c = 3
			value[gi] = u
			cu := *u
			check[gi] = &cu
			sema <- struct{}{}
		}(i)
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		if !reflect.DeepEqual(value[i], check[i]) {
			t.Fatalf("%d value=%+v , check =%+v", i, value[i], check[i])
		}
	}
	a.Free()
}

func BenchmarkAlloc_3int_P1(b *testing.B) {
	a := NewArena(unsafe.Sizeof(user{}))
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a.Alloc()
	}
	a.Free()
	runtime.GC()
}

func BenchmarkAlloc_3int(b *testing.B) {
	a := NewArena(unsafe.Sizeof(user{}))
	b.SetBytes(9)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Alloc()
		}
	})
	a.Free()
	runtime.GC()
}
