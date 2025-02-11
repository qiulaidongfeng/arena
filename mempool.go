package arena

import (
	"sync"
	"sync/atomic"
)

type MemPool[T any] struct {
	// 互斥锁仅用于同步内存池自动扩容
	lock sync.Mutex
	// 注意不应该直接访问 bufs
	// 而应该通过bufas访问，例外是访问创建后只读的内容
	// 这样大部分情况不用互斥锁
	// 仅用两个 sync/atomic.LoadPointer
	//
	// 互斥锁的性能，因为多个goroutine竞争同一块内存的原子操作，随CPU核数增加反而减小
	// 原子读（sync/atomic.LoadPointer）的性能即使同样面对，多个goroutine竞争同一块内存的原子操作，还是非常高
	// 因为读操作和读操作可以同时进行，写操作不能同时和其他读或写操作同时进行
	//
	// 这里故意不用读写锁，因为memPool唯一需要同步的时候是
	// bufs = append(bufs, nbuf)
	// 互斥锁够用了，内存池自动扩容时本来就只能没办法利用多核
	// 这样内存池不扩容时避免引入原子加操作，导致激烈的竞争对同一块内存进行原子加操作，引发多核性能降低
	bufs []*buf[T]
	// 每个go类型T的 *runtime._type
	rtype uintptr
	//bufas 是内存池当前正在使用的一段连续内存
	bufas atomic.Pointer[buf[T]]
	// bufsize 是自动扩容分配的内存，默认为64Mb
	bufsize int64
}

// NewMemPool 创建一组代表一起分配和释放的Go值的内存池
//
//   - dataSize 是单个Go值大小，单位是字节
//   - rtype 是go类型T的 *runtime._type
//   - bufsize 是可选的，决定 [MemPool] 每次自动扩容分配多大内存，默认为64Mb，必须大于dataSize
func NewMemPool[T any](dataSize uintptr, rtype uintptr, bufsize ...int64) *MemPool[T] {
	ret := &MemPool[T]{rtype: rtype}
	var nbuf *buf[T]
	if len(bufsize) == 0 {
		ret.bufsize = defaultmem
		nbuf = newbuf[T](defaultmem, rtype, int64(dataSize))
	} else {
		ret.bufsize = bufsize[0]
		nbuf = newbuf[T](bufsize[0], rtype, int64(dataSize))
	}
	bufs := make([]*buf[T], 1)
	bufs[0] = nbuf
	ret.bufs = bufs
	ret.bufas.Store(nbuf)
	return ret
}

// reAlloc 执行扩容
func (a *MemPool[T]) reAlloc() {
	a.lock.Lock()
	nbuf := newbuf[T](a.bufsize, a.rtype, a.bufs[0].dataSize)
	a.bufs = append(a.bufs, nbuf)
	a.bufas.Store(nbuf)
	a.lock.Unlock()
}

// Alloc 分配一个Go值，并返回指针
func (a *MemPool[T]) Alloc() *T {
	buf := a.bufas.Load()
	return buf.move(a)
}

// AllocSlice 分配连续的go值，并返回切片
func (a *MemPool[T]) AllocSlice(len, cap int) []T {
	buf := a.bufas.Load()
	return buf.moveSlice(a, len, cap)
}

// Free 释放所有Go值
//
// 从调用的一刻开始，通过 [MemPool.Alloc] 或 [MemPool.AllocSlice] 获取的Go值，读写内存，行为是未定义且不安全的
func (a *MemPool[T]) Free() {
	for i := 0; i < len(a.bufs); i++ {
		a.bufs[i].Free()
	}
}
