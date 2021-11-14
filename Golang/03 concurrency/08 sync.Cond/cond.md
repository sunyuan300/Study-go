# Cond条件变量
条件变量用来协调那些想要访问共享资源的goroutine，当共享资源的状态发生变化的时候，它可以用来通知被互斥锁阻塞的goroutine。

经常用在多个goroutine等待，一个goroutine通知（事件发生）的场景。如果是一个通知，一个等待，使用互斥锁或channel就能搞定了。

**场景**:

一个协程在异步地接收数据，其它多个协程必须等待这个协程接收完数据，才能读取到正确的数据。在这种情况下，如果单纯使用channel或互斥锁，只能有一个协程可以等待，并读取到数据，没办法通知其他的协程也读取数据。

这个时候，就需要有个全局的变量来标志第一个协程数据是否接受完毕，剩下的协程，反复检查该变量的值，直到满足要求。或者创建多个 channel，每个协程阻塞在一个 channel 上，由接收数据的协程在数据接收完毕后，逐个通知。总之，需要额外的复杂度来完成这件事。

# 结构体
```go
// Each Cond has an associated Locker L (Often a *Mutex or  *RWMutex),Which must be held when changing the condition and when calling the Wait method.
// 每个Cond实例都会关联一个锁L(通常是*Mutex或*RWMutex),当修改条件或调用Wait方法时,必须加锁。
type Cond struct {
        noCopy noCopy
        L Locker
        notify  notifyList
        checker copyChecker
}
```
- noCopy：用于保证结构体不会在编译期间拷贝。
- copyChecker：用于禁止运行期间发生的拷贝。
- L：用于保护内部的 notify 字段，Locker 接口类型的变量。
- notify：一个Goroutine的链表，是实现同步机制的核心结构。
```go
type notifyList struct {
	wait uint32
	notify uint32

	lock mutex
	head *sudog
	tail *sudog
}
```
`head` 和 `tail` 分别指向的链表的头和尾，`wait` 和 `notify` 分别表示当前正在等待的和已经通知到的 Goroutine索引。

# 接口
- `func NewCond(l Locker) *Cond`：创建 Cond 实例时，需要关联一个锁。
- `Broadcast`：唤醒所有陷入休眠的goroutine，若没有陷入休眠的goroutine，也不会报错
- `Signal`：唤醒队列最前面的Goroutine。
- `Wait`：将当前goroutine陷入休眠状态，并加入通知队列。Unlock()->***阻塞等待通知(即等待Signal()或Broadcast()的通知)->收到通知***->Lock()

# Wait
```go
func (c *Cond) Wait() {
	c.checker.check()
	t := runtime_notifyListAdd(&c.notify) //runtime.notifyListAdd 的链接名
	c.L.Unlock()    // 解锁
	runtime_notifyListWait(&c.notify, t)    // 阻塞等待通知
	c.L.Lock()  // 加锁
}

func notifyListAdd(l *notifyList) uint32 {
	return atomic.Xadd(&l.wait, 1) - 1
}
```
1. 调用`runtime.notifyListAdd`将等待计数器加一并解锁。
2. 调用`runtime.notifyListWait`等待其它goroutine的唤醒并加锁。


runtime.notifyListWait会获取当前Goroutine并将它追加到Goroutine通知链表的最末端。同时还会调用`runtime.goparkunlock`将当前Goroutine陷入休眠，该函数也是在Go语言切换Goroutine时经常会使用的方法，它会直接让出当前处理器的使用权并等待调度器的唤醒。
```go
func notifyListWait(l *notifyList, t uint32) {
	s := acquireSudog()
	s.g = getg()
	s.ticket = t
	if l.tail == nil {
		l.head = s
	} else {
		l.tail.next = s
	}
	l.tail = s
	goparkunlock(&l.lock, waitReasonSyncCondWait, traceEvGoBlockCond, 3)
	releaseSudog(s)
}
```
         
## 例子1
```go
package main

import (
	"log"
	"sync"
	"time"
)

var done = false

func read(name string, c *sync.Cond) {
	c.L.Lock()
	for !done {
		c.Wait()
	    log.Println(name, "receive signal")
	}
	log.Println(name, "starts reading")
	c.L.Unlock()
}

func write(name string, c *sync.Cond) {
	log.Println(name, "starts writing")
	time.Sleep(time.Second)
	c.L.Lock()
	done = true
	c.L.Unlock()
	log.Println(name, "wakes all")
	c.Broadcast()
}

func main() {
	cond := sync.NewCond(&sync.Mutex{})

	go read("reader1", cond)
	go read("reader2", cond)
	go read("reader3", cond)
	write("writer", cond)

	time.Sleep(time.Second * 3)
}

/* output:
    2021/11/09 16:32:19 writer starts writing
    2021/11/09 16:32:20 writer wakes all
    2021/11/09 16:32:20 reader3 receive signal
    2021/11/09 16:32:20 reader3 starts reading
    2021/11/09 16:32:20 reader1 receive signal
    2021/11/09 16:32:20 reader1 starts reading
    2021/11/09 16:32:20 reader2 receive signal
    2021/11/09 16:32:20 reader2 starts reading
*/
```

- done:互斥锁需要保护的**条件变量**。
- read():调用wait()等待*Signal*或*Broadcast*通知,直到条件变量done变为true。
- write():将done置为true,调用*Broadcast*通知所有等待的goroutine。

## 例子2
```go
var status int64

func main() {
	c := sync.NewCond(&sync.Mutex{})
	for i := 0; i < 10; i++ {
		go listen(c)
	}
	time.Sleep(1 * time.Second)
	go broadcast(c)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func broadcast(c *sync.Cond) {
	c.L.Lock()
	atomic.StoreInt64(&status, 1)
	c.Broadcast()
	c.L.Unlock()
}

func listen(c *sync.Cond) {
	c.L.Lock()
	for atomic.LoadInt64(&status) != 1 {
		c.Wait()
	}
	fmt.Println("listen")
	c.L.Unlock()
}
```

上述代码同时运行了11个Goroutine，这11个Goroutine分别做了不同事情：
- 10个Goroutine通过Cond.Wait等待特定条件的满足。
- 1个Goroutine调用Cond.Broadcast唤醒所有陷入等待的Goroutine。

![](cond.png)

# 总结
一般情况下，我们都会先调用Cond.Wait陷入休眠**等待满足期望条件**，当满足唤醒条件时，就可以选择使用 Cond.Signal或者Cond.Broadcast唤醒一个或者全部的goroutine。


# Reference
[同步原语与锁](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#cond)

[读写锁和互斥锁的性能比较](https://geektutu.com/post/hpg-mutex.html)

[条件变量](https://cyent.github.io/golang/goroutine/sync_cond/)