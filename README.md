# arena

[English](./README_en.md)

#### 介绍
go1.20引入了arena.Arena，但不能同时被多个goroutine使用,且go1.n(n<20)不能使用，本包提供解除这些限制的Arena

#### 特点
   - 可以同时在多个goroutine使用
   - 可以在版本低于go1.20使用 (>=go1.18)
   - 不会发生释放后使用

#### module path and import path

应该这样使用

```go
import arenas "gitee.com/qiulaidongfeng/arena"
```

或

```go
import arenas "github.com/qiulaidongfeng/arena"
```
