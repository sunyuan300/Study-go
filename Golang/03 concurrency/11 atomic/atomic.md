# 原子操作
Golang锁机制的底层是基于原子操作的，一般是通过CPU指令实现。Golang中的原子操作由`sync/atomic`提供。

# 函数签名
1. LoadXXX：从某个地址中取值(*读操作*)
```go
func LoadInt32(addr *int32) (val int32)
func LoadInt64(addr *int64) (val int64)
func LoadUint32(addr *uint32) (val uint32)
func LoadUint64(addr *uint64) (val uint64)
func LoadUintptr(addr *uintptr) (val uintptr)
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)
```

2. StoreXXX：给某个地址赋值(*写操作*)
```go
func StoreInt32(addr *int32, val int32)
func StoreInt64(addr *int64, val int64)
func StoreUint32(addr *uint32, val uint32)
func StoreUint64(addr *uint64, val uint64)
func StoreUintptr(addr *uintptr, val uintptr)
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
```

3. AddXXX：给某个值加上delta(*修改操作*)
```go
func AddInt32(addr *int32, delta int32) (new int32)
func AddInt64(addr *int64, delta int64) (new int64)
func AddUint32(addr *uint32, delta uint32) (new uint32)
func AddUint64(addr *uint64, delta uint64) (new uint64)
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr)
```

4. SwapXXX：交互两个值，并返回旧值(*交换操作*)
```go
func SwapInt32(addr *int32, new int32) (old int32)
func SwapInt64(addr *int64, new int64) (old int64)
func SwapUint32(addr *uint32, new uint32) (old uint32)
func SwapUint64(addr *uint64, new uint64) (old uint64)
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr)
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer)
```

5. CompareAndSwapXXX CAS操作：先比较传入的地址值是否等于old，如果相等则赋于新值，如果相等的话就直接返回 false，返回 true 时表示赋值成功。(*比较并交换操作*)
```go
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)
```

6. `Value`用于任意类型值的Store、Load。
以上函数只能支持几种基本数据类型，为了扩大原子操作的范围，Go1.4引入了`Value`类型。
```go
type Value struct {v interface{}}
func (v *Value) Load() (x interface{})
func (v *Value) Store(x interface{})
```

# CAS
在`sync/atomic`包中除了`Value`之外，其他函数都没有直接的源码，需要去`runtime/internal/atomic`中查找，这里以`CAS`函数为例。

```go
// bool Cas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕atomic·Cas(SB),NOSPLIT,$0-17
	MOVQ	ptr+0(FP), BX
	MOVL	old+8(FP), AX
	MOVL	new+12(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	SETEQ	ret+16(FP)
	RET
```
看这个具体汇编代码就可以发现，使用了 LOCK 来保证操作的原子性，`CMPXCHG`指令其实就是 CPU 实现的 CAS 操作。
>关于 LOCK 指令通过查阅 intel 的手册可以发现，对于P6之前的处理器，LOCK 指令会总是锁总线，但是 P6 之后可能会执行“缓存锁定”，如果被锁定的内存区域被缓存在了处理器中，这个时候会通过缓存一致性来保证操作的原子性。

# Value

## 结构体
```go
type Value struct {
	v interface{}
}
```
## Store
将原始的变量x存放到一个atomic.Value类型的v里。
```go
func (v *Value) Store(x interface{}) {
	if x == nil {
		panic("sync/atomic: store of nil value into Value")
	}
    // ifaceWords 其实就是定义了一下 interface 的结构，包含 data 和 type 两部分
    // 这里 vp 是原有值
    // xp 是传入的值
	vp := (*ifaceWords)(unsafe.Pointer(v))
	xp := (*ifaceWords)(unsafe.Pointer(&x))
    // for 循环不断尝试
	for {
        // 这里先用原子方法取一下老的类型值
		typ := LoadPointer(&vp.typ)
		if typ == nil {
            // 等于 nil 就说明这是第一次 store
            // 调用 runtime 的方法禁止抢占，避免操作完成一半就被抢占了
            // 同时可以避免 GC 的时候看到 unsafe.Pointer(^uintptr(0)) 这个中间状态的值
			runtime_procPin()
			if !CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(^uintptr(0))) {
				runtime_procUnpin()
				continue
			}

			// 分别把值和类型保存下来
			StorePointer(&vp.data, xp.data)
			StorePointer(&vp.typ, xp.typ)
			runtime_procUnpin()
			return
		}

		if uintptr(typ) == ^uintptr(0) {
            // 如果判断发现这个类型是这个固定值，说明当前第一次赋值还没有完成，所以进入自旋等待
			continue
		}
		// 第一次赋值已经完成，判断新的赋值的类型和之前是否一致，如果不一致就直接 panic
		if typ != xp.typ {
			panic("sync/atomic: store of inconsistently typed value into Value")
		}
        // 保存值
		StorePointer(&vp.data, xp.data)
		return
	}
}
```
## Load
从线程安全的v中读取上一步存放的内容。
```go
func (v *Value) Load() (x interface{}) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
    // 先拿到类型值
	typ := LoadPointer(&vp.typ)
    // 这个说明还没有第一次 store 或者是第一次 store 还没有完成
	if typ == nil || uintptr(typ) == ^uintptr(0) {
		// First store not yet completed.
		return nil
	}
    // 获取值
	data := LoadPointer(&vp.data)
    // 构造 x 类型
	xp := (*ifaceWords)(unsafe.Pointer(&x))
	xp.typ = typ
	xp.data = data
	return
}
```

# 总结
虽然在一些情况下 atomic 的性能要好很多，但是这个是一个 low level 的库，在实际的业务代码中最好还是使用 channel 但是我们也需要知道，在一些基础库，或者是需要极致性能的地方用上这个还是很爽的，但是使用的过程中一定要小心，不然还是会容易出 bug。

# Reference
[深入理解 sync/atomic](https://lailin.xyz/post/go-training-week3-atomic.html)

[Go 语言标准库中 atomic.Value 的前世今生](https://blog.betacat.io/post/golang-atomic-value-exploration/)