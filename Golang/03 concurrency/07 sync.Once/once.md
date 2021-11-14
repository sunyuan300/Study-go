# 函数签名
```go
type Once struct{...}
func (o *Once) Do(f func())
```

# 源码分析

## Once
```go
type Once struct {
	// done indicates whether the action has been performed.
	// It is first in the struct because it is used in the hot path.
	// The hot path is inlined at every call site.
	// Placing done first allows more compact instructions on some architectures (amd64/386),
	// and fewer instructions (to calculate offset) on other architectures.
	done uint32
	m    Mutex
}
```
- done：用于表示函数是否已执行，0表示还未执行，1表示已执行，会直接返回。
- m：互斥锁。

## Do

```go
func (o *Once) Do(f func()) {
	// Note: Here is an incorrect implementation of Do:
	//
	//	if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
	//		f()
	//	}
	//
	// Do guarantees that when it returns, f has finished.
	// This implementation would not implement that guarantee:
	// given two simultaneous calls, the winner of the cas would
	// call f, and the second would return immediately, without
	// waiting for the first's call to f to complete.
	// This is why the slow path falls back to a mutex, and why
	// the atomic.StoreUint32 must be delayed until after f returns.

	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}
```
1. 为当前 Goroutine 获取互斥锁；
2. 执行传入的无入参函数f()；
3. 运行延迟函数调用，将成员变量 done 更新成 1；

## 关于Do的错误实现
官方文档中，给出了一种Do方法的错误实现：
```go
if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
		f()
    }
```

Do方法的设计原则是：f()执行完成后，才会返回。

但是上述实现却不能保证，原因是：当多个goroutine并发的调用Do方法时，第一个使用CAS方法获取到锁的goroutine会执行f()，但是其它goroutine不会等待第一个goroutine执行完f()，而是直接返回。

# 例子1
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once
	onceBody := func() {
		fmt.Println("Only once")
	}
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			once.Do(onceBody)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
//Output: Only once
```

## 例子2
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once
	f1 := func() {
		fmt.Println("f1")
	}

	f2 := func() {
		fmt.Println("f2")
	}
	
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			once.Do(f1)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			once.Do(f2)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

//Output: f1或f2
```

# 总结
- Once 保证了传入的函数只会执行一次，常用在单例模式、配置文件加载、初始化等场景。
- Once.Do 方法即使多次传入不同的函数，也只会执行第一次调传入的函数；
-  Once.Do 方法中传入的函数只会被执行一次，即使函数中发生了panic。

# Reference
[Once](https://pkg.go.dev/sync#Once.Do)

[Go并发编程(八) 深入理解 sync.Once](https://lailin.xyz/post/go-training-week3-once.html)

[6.2 同步原语与锁](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#once)