package arena

import (
	"sync/atomic"
	"unsafe"
)

type buf struct {
	buf       []byte
	index     int64
	max_index int64
	dataSize  int64
}

func newbuf(dataSize int64, bufsize int64) *buf {
	datan := bufsize / dataSize
	return &buf{
		buf:       make([]byte, datan*dataSize, datan*dataSize),
		max_index: int64(datan) * dataSize,
		index:     0,
		dataSize:  dataSize,
	}
}

// move 从切片中分配固定大小的Go值，切片容量不足时会自动扩容
func (b *buf) move(a *Arena) unsafe.Pointer {
	n_index := atomic.AddInt64(&b.index, b.dataSize)
	if n_index >= b.max_index { //切片容量不足
		a.reAlloc()      //扩容
		return a.Alloc() //重新分配
	}
	return unsafe.Pointer(&b.buf[n_index])
}

// empty 判断切片是否有可分配的部分
func (b *buf) empty() bool {
	return atomic.LoadInt64(&b.index) >= b.max_index
}

func (b *buf) Free() {
	b.buf = nil
}
