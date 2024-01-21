package arena

func ExampleArena() {
	type user struct {
		a, b, c int
	}
	a := New()
	//分配并使用
	b := Alloc[user](a)
	b.a = 1
	b.c = 2
	//使用完毕
	a.Free()
}
