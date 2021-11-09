```go
package main

func main()  {
 
}
```

## sync包的使用
sync包主要用于并发同步，整个包都是围绕`Locker`进行的。***该包下的对象，在使用过之后，千万不要复制***。

其下的atomic包，提供了一些底层的**原子操作**。

```go
type Locker interface {
    Lock()
    Unlock()
}
```

## 互斥锁Mutex
- 一个互斥锁只能同时被一个 goroutine 锁定，其它 goroutine 将阻塞直到互斥锁被解锁。
- 对一个未锁定的互斥锁进行解锁将会产生运行时错误。
```go
package main

import "sync"

func main()  {
	mux := sync.Mutex{}
	mux.Unlock()
	//fatal error: sync: unlock of unlocked mutex
}
```

## 读写锁RWMutex
与互斥锁最大的不同就是可以分别对`读`、`写`进行锁定。通常用在***读多写少***的场景。
```go
func (rw *RWMutex) Lock()       //对写加锁
func (rw *RWMutex) Unlock()     //对写解锁

func (rw *RWMutex) RLock()      //对读加锁
func (rw *RWMutex) RUnlock()    //对读解锁
```

- 读锁之间不互斥，没有写锁的情况下，读锁是无阻塞的，多个协程可以同时获得读锁。
- 写锁之间是互斥的，存在写锁，其他写锁阻塞。
- 写锁与读锁是互斥的，如果存在读锁，写锁阻塞，如果存在写锁，读锁阻塞。

## Cond条件变量
条件变量用来协调那些想要访问共享资源的goroutine，当共享资源的状态发生变化的时候，它可以用来通知被互斥锁阻塞的goroutine。

经常用在多个goroutine等待，一个goroutine通知（事件发生）的场景。如果是一个通知，一个等待，使用互斥锁或channel就能搞定了。

**场景**:

一个协程在异步地接收数据，其它多个协程必须等待这个协程接收完数据，才能读取到正确的数据。在这种情况下，如果单纯使用channel或互斥锁，只能有一个协程可以等待，并读取到数据，没办法通知其他的协程也读取数据。

这个时候，就需要有个全局的变量来标志第一个协程数据是否接受完毕，剩下的协程，反复检查该变量的值，直到满足要求。或者创建多个 channel，每个协程阻塞在一个 channel 上，由接收数据的协程在数据接收完毕后，逐个通知。总之，需要额外的复杂度来完成这件事。

### Cond的四个方法
```go
// Each Cond has an associated Locker L (Often a *Mutex or  *RWMutex),Which must be held when changing the condition and when calling the Wait method.
// 每个Cond实例都会关联一个锁L(通常是*Mutex或*RWMutex),当修改条件或调用Wait方法时,必须加锁。
type Cond struct {
        noCopy noCopy

        // L is held while observing or changing the condition
        L Locker

        notify  notifyList
        checker copyChecker
}
```

- `func NewCond(l Locker) *Cond`：创建 Cond 实例时，需要关联一个锁。
- `Broadcast`：通知所有Wait()了的goroutine，若没有Wait()，也不会报错
- `Signal`：通知一个Wait()了的goroutine，若没有Wait()，也不会报错。Signal()通知的顺序是根据原来加入通知列表(Wait())的先入先出。
- `Wait`：Unlock()->***阻塞等待通知(即等待Signal()或Broadcast()的通知)->收到通知***->Lock()

```go
func (c *Cond) Wait() {
	c.checker.check()
	t := runtime_notifyListAdd(&c.notify)
	c.L.Unlock()    // 加锁
	runtime_notifyListWait(&c.notify, t)    // 阻塞等待通知
	c.L.Lock()  // 加锁
}
```
         
### 例子
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




## Reference
[读写锁和互斥锁的性能比较](https://geektutu.com/post/hpg-mutex.html)

[条件变量](https://cyent.github.io/golang/goroutine/sync_cond/)
