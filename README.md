# arena

#### 介绍
go1.20引入了arena.Arena，但不能同时被多个goroutine使用,且go1.n(n<20)不能使用，本包提供解除这些限制的Arena

#### 特点
   - 可以同时在多个goroutine使用
   - 不使用泛型，可以在版本低于go1.20使用
   - 只能分配固定大小的Go值

