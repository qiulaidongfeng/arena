package arena

import (
	"sync/atomic"
	"unsafe"
)

type buf struct {
	buf       []byte
	index     uintptr
	max_index uintptr
}

func newbuf(bufsize int64) *buf {
	return &buf{
		buf:       make([]byte, bufsize, bufsize),
		max_index: uintptr(bufsize),
		index:     0,
	}
}

// move 从切片中分配固定大小的Go值，切片容量不足时会自动扩容
func (b *buf) move(a *Arena, dataSize uintptr) unsafe.Pointer {
	n_index := atomic.AddUintptr(&b.index, dataSize)
	if n_index >= b.max_index { //切片容量不足
		a.reAlloc()              //扩容
		return a.Alloc(dataSize) //重新分配
	}
	return unsafe.Pointer(&b.buf[n_index])
}

// empty 判断切片是否有可分配的部分
func (b *buf) empty() bool {
	return atomic.LoadUintptr(&b.index) >= b.max_index
}

func (b *buf) Free() {
	b.buf = nil
}
