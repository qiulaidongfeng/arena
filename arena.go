package arena

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const defaultmem = 1024 * 1024 * 8 //8mb

// Arena是go1.20引入的arena.Arena的升级替代，创建后不得复制
//
// 与go1.20引入的arena.Arena相比，有下列不同
//   - 可以同时在多个goroutine使用
//   - 不使用泛型，可以在版本低于go1.20使用
//   - 只能分配固定大小的Go值
type Arena struct {
	lock    sync.Mutex
	bufas   atomic.Value //type=*bufsi
	bufsize int64
}

type bufsi struct {
	bufs []*buf
	i    int64
}

// NewArena 创建一组代表一起分配和释放的Go值的集合，内存大小为n*dataSize（n=bufsize/dataSize）
//   - dataSize 是单个Go值大小，决定 [Arena.Alloc] 返回的指针指向的内存的长度，单位是字节
//   - bufsize 是可选的，决定 [Arena.Alloc] 每次自动扩容分配多大内存，默认为8Mb,必须大于dataSize
func NewArena(dataSize uintptr, bufsize ...int64) *Arena {
	ret := &Arena{}
	var nbuf *buf
	if len(bufsize) == 0 {
		ret.bufsize = defaultmem
		nbuf = newbuf(int64(dataSize), defaultmem)
	} else {
		ret.bufsize = bufsize[0]
		nbuf = newbuf(int64(dataSize), bufsize[0])
	}
	bufs := make([]*buf, 1)
	bufs[0] = nbuf
	bufas_value := &bufsi{bufs: bufs, i: 0}
	ret.bufas.Store(bufas_value)
	return ret
}

// reAlloc 执行扩容
func (a *Arena) reAlloc() {
	a.lock.Lock()
	bufs := a.bufas.Load().(*bufsi)
	if bufs.bufs[bufs.i].empty() {
		bufsn := append(bufs.bufs, newbuf(bufs.bufs[0].dataSize, a.bufsize))
		a.bufas.Store(&bufsi{bufs: bufsn, i: bufs.i + 1})
	}
	a.lock.Unlock()
}

// Alloc分配一个固定大小的Go值，并返回指针，[NewArena] 的dataSize是这个固定大小
//
// 返回的指针如果访问超过固定大小的内存，行为是未定义且不安全的
func (a *Arena) Alloc() unsafe.Pointer {
	bufs := a.bufas.Load().(*bufsi)
	return bufs.bufs[bufs.i].move(a)
}

// Free 释放所有Go值
//
// 从调用的一刻开始，通过 [Arena.Alloc] 获取的指针，读写内存，行为是未定义且不安全的
func (a *Arena) Free() {
	bufs := a.bufas.Load().(*bufsi)
	for i := 0; i < len(bufs.bufs); i++ {
		bufs.bufs[bufs.i].Free()
	}
}
