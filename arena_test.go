package arena

import (
	"math"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
)

const n = 100000

type user struct {
	a int
	b int
	c int
}

type userptr struct {
	a *int
	b *int
	c *int
}

func TestAlloc(t *testing.T) {
	var value, check []*user = make([]*user, n), make([]*user, n)
	a := New()
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
			u := Alloc[user](a)
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

func BenchmarkAlloc_int_P1(b *testing.B) {
	a := New()
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Alloc[int](a)
	}
	a.Free()
	runtime.GC()
	debug.FreeOSMemory()
}

func BenchmarkAlloc_int(b *testing.B) {
	a := New()
	b.SetBytes(9)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Alloc[int](a)
		}
	})
	a.Free()
	runtime.GC()
	debug.FreeOSMemory()
}

func TestAllocPtr(t *testing.T) {
	var value, check []*userptr = make([]*userptr, n), make([]*userptr, n)
	a := New()
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
			u := Alloc[userptr](a)
			u.a = new(int)
			u.b = new(int)
			u.c = new(int)
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

func TestAllocSliceint_Len2(t *testing.T) {
	var value, check [][]int = make([][]int, n), make([][]int, n)
	a := New()
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
			u := AllocSlice[int](a, 2, 2)
			u[0] = math.MaxInt64
			u[1] = math.MaxInt64
			value[gi] = u
			cu := make([]int, 2)
			cu[0] = math.MaxInt64
			cu[1] = math.MaxInt64
			check[gi] = cu
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

func BenchmarkAlloc_int_new(b *testing.B) {
	f := new(int)
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f = new(int)
	}
	*f = 1
	runtime.GC()
	debug.FreeOSMemory()
}

func BenchmarkAlloc_intAndFree(b *testing.B) {
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := New()
		for i := 0; i < 100; i++ {
			Alloc[int](a)
		}
		a.Free()
	}
}

func TestMain(m *testing.M) {
	println("UseAfterFree(false)")
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
	UseAfterFree(true)
	println("UseAfterFree(true)")
	os.Exit(m.Run())
}

func TestAllocSoonUse(t *testing.T) {
	var value, check []*user = make([]*user, n), make([]*user, n)
	a := New()
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
			u := AllocSoonUse[user](a)
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

func BenchmarkAllocSoonUse_int_P1(b *testing.B) {
	a := New()
	b.SetBytes(9)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AllocSoonUse[int](a)
	}
	a.Free()
	runtime.GC()
	debug.FreeOSMemory()
}
