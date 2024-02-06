package arena

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const defaultmem = 8 * 1024 * 1024 //8mb

// Arena是go1.20引入的arena.Arena的升级替代
//
// 它通过为每个类型准备内存池实现
//
// 与go1.20引入的arena.Arena相比，有下列不同
//   - 可以同时在多个goroutine使用
//   - 可以在版本低于go1.20使用
//   - 不会发生释放后使用
//
// 创建后不得复制
type Arena struct {
	typs map[uintptr]interface{} //value=*memPool
	lock sync.RWMutex
}

// New创建一个 [*Arena]
func New() *Arena {
	return &Arena{
		typs: make(map[uintptr]interface{}),
	}
}

// getMemPool 获取 [*Arena] 中某个类型的 [*memPool]
func getMemPool[T any](a *Arena, bufsize ...int64) *MemPool[T] {
	var zero T
	rtype := rtypeOf(zero)
	a.lock.RLock()
	// 小心这里的想法未来成为现实
	// https://github.com/golang/go/issues/62483#issuecomment-1800913220
	// Can we put all the runtime._types in one of these weak maps?
	// So when no more pointers to that runtime._type exist, the entry itself can be garbage collected.
	// 导致同一类型的*rtype不同
	mempool := a.typs[rtype]
	if mempool == nil {
		a.lock.RUnlock()
		a.lock.Lock()
		mempool = NewMemPool[T](unsafe.Sizeof(zero), rtype, bufsize...)
		a.typs[rtype] = mempool
		a.lock.Unlock()
	} else {
		a.lock.RUnlock()
	}
	return mempool.(*MemPool[T])
}

// Alloc 从 [*Arena] 中创建一个 T , 并返回指针
//
//   - bufsize 是可选的，决定 [MemPool.Alloc] 每次自动扩容分配多大内存，默认为8Mb,必须大于unsafe.Sizeof(*new(T))
func Alloc[T any](a *Arena, bufsize ...int64) *T {
	mempool := getMemPool[T](a, bufsize...)
	return mempool.Alloc()
}

// Alloc 从 [*Arena] 中创建一些连续的 T , 并返回切片
//
//   - bufsize 是可选的，决定 [MemPool.AllocSlice] 每次自动扩容分配多大内存，默认为8Mb,必须大于unsafe.Sizeof(*new(T))*uintptr(cap)
func AllocSlice[T any](a *Arena, len, cap int, bufsize ...int64) []T {
	mempool := getMemPool[T](a, bufsize...)
	return mempool.AllocSlice(len, cap)
}

// AllocSoonUse 从 [*Arena] 中创建一个 T , 并返回指针
//
//   - bufsize 是可选的，决定 [MemPool.Alloc] 每次自动扩容分配多大内存，默认为8Mb,必须大于unsafe.Sizeof(*new(T))
//
// 对这个函数的多次调用将尽可能把返回的指针指向接近的内存地址，以提高很快就要使用的一批不同类型数据的缓存命中率
//
// 请勿从此函数分配含有指针的类型，否则分配出的内存将对GC隐藏这个类型的对象中的指针，这可能导致一些还在使用的内存被GC回收
func AllocSoonUse[T any](a *Arena, bufsize ...int64) *T {
	mempool := getMemPool[byte](a, bufsize...)
	size := int(unsafe.Sizeof(*new(T)))
	buf := mempool.AllocSlice(size, size)
	return (*T)(unsafe.Pointer(&buf[0]))
}

// Free 释放所有内存，只能调用一次，否则行为未定义
func (a *Arena) Free() {
	// Free 假设Arena用完后，只会调用1次Free,所以不加锁
	for _, v := range a.typs {
		v.(interface {
			Free()
		}).Free()
	}
}

var enableUseAfterFree int64

// EnableUseAfterFree 使得 [*Arena] 和 [*MemPool] 可以发生
// use-after-free (释放后使用) 来优化性能
//
// 因为只要有一个 [*Arena] 或 [*MemPool] 可以发生
// use-after-free (释放后使用) 就不能保证不发生数据竞争等问题
// 所以设计成 use-after-free (释放后使用)
// 不是不会发生就是都有可能发生
func UseAfterFree(enable bool) {
	if enable {
		atomic.StoreInt64(&enableUseAfterFree, 1)
		return
	}
	atomic.StoreInt64(&enableUseAfterFree, 0)
}
