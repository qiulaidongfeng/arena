package arena

import (
	"unsafe"
)

func ExampleArena() {
	type user struct {
		a, b, c int
	}
	a := NewArena(unsafe.Sizeof(user{}))
	//分配并使用
	b := (*user)(a.Alloc())
	b.a = 1
	b.c = 2
	//使用完毕
	a.Free()
}
