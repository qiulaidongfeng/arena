package arena

import (
	"sync/atomic"
	"unsafe"
)

// buf 持有一段连续的go值，与 [MemPool] 合作实现内存池
type buf[T any] struct {
	buf      []T
	index    int64
	dataSize int64
}

func newbuf[T any](bufsize int64, rtype uintptr, dataSize int64) *buf[T] {
	return &buf[T]{
		buf: newSlice[T](
			runtime_mallocgc(uintptr(bufsize), rtype, true),
			bufsize,
			dataSize),
		index:    0,
		dataSize: dataSize,
	}
}

// 使用这个类型的代码
// 依赖go的切片是下面的布局
type slice struct {
	ptr      unsafe.Pointer
	len, cap int
}

// newSlice 创建一个切片，从ptr开始引用一段内存
func newSlice[T any](ptr unsafe.Pointer, cap int64, dataSize int64) (ret []T) {
	p := (*slice)(unsafe.Pointer(&ret))
	p.ptr = ptr
	p.cap = int(cap / dataSize)
	p.len = p.cap
	return
}

// move 从切片中分配固定大小的Go值，切片容量不足时会自动扩容
func (b *buf[T]) move(a *MemPool[T]) *T {
	n_index := atomic.AddInt64(&b.index, 1)
	// 这是为了支持moveSlice
	// 还避免浪费切片开头的1个Go值
	// 从基准测试看，只让单线程性能变慢了<1ns
	// 作者认为值得
	n_index -= 1
	if n_index >= int64(len(b.buf)) { //切片容量不足
		a.reAlloc()      //扩容
		return a.Alloc() //重新分配
	}
	return &b.buf[n_index]
}

// moveSlice 从切片中分配多个固定大小的Go值，切片容量不足时会自动扩容
func (b *buf[T]) moveSlice(a *MemPool[T], Len int, cap int) []T {
	// add_index 是需要分配的Go值数量
	add_index := int64(cap)
	n_index := atomic.AddInt64(&b.index, add_index)
	if n_index > int64(len(b.buf)) { //切片容量不足
		a.reAlloc()                   //扩容
		return a.AllocSlice(Len, cap) //重新分配
	}
	return ptrToSlice[T](unsafe.Pointer(&b.buf[n_index-add_index]), Len, cap)
}

// ptrToSlice 创建一个切片，从ptr开始引用一段内存
func ptrToSlice[T any](ptr unsafe.Pointer, len, cap int) (ret []T) {
	p := (*slice)(unsafe.Pointer(&ret))
	p.ptr = ptr
	p.len = len
	p.cap = cap
	return ret
}

// Free 释放内存池的内存
func (b *buf[T]) Free() {
	b.buf = nil
}

//go:linkname runtime_mallocgc runtime.mallocgc
func runtime_mallocgc(size uintptr, typ uintptr, needzero bool) unsafe.Pointer

// rtypeOf directly extracts the *rtype of the provided value.
func rtypeOf(i any) uintptr {
	// 这里依赖go的接口相当于
	// type eface struct {
	// typ *runtime._type
	// data unsafe.Pointer
	//	}
	eface := (*[2]uintptr)(unsafe.Pointer(&i))
	return eface[0]
}
