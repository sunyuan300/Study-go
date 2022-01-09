项目介绍:
- 背景:
- 职责:
- 技术难点、挑战:
- 解决方案：为什么这么做？为什么用A不用B?


输入网址到显示画面的详细流程。

Golang相关:
1. new和make的使用场景?
2. 函数结构体,传值还是指针？
3. GMP模型
   - 线程有哪几种模型？
   - goroutine什么时候会发生阻塞？  
   - goroutine有哪几种状态？线程呢？
   - 如果一个Goroutine一直占用资源,GMP模型是怎么解决这个问题的？
   - 如果若干线程中一个线程OOM,会发生什么？如果是goroutine呢？
4. 启动若干个Goroutine,其中一个panic,会发生什么?
5. 项目中错误/日志是怎么处理的？

6. Gin框架怎么做参数校验？
7. 中间件使用过吗？

8. Channel
   - 对已关闭的Channel写数据会怎样？
    
9. 锁
   - Mutex:有哪几种模式？底层实现？

```go
func main () {
	var out []*int
	for i := 0; i < 3; i++ {
		out = append(out,&i)
	}
    fmt.Println(*out[0],*out[1],*out[2])
}
```

```go
 func main() {
     x := []int{1,2,3,4,5}
     y := x[1:4]
     y = append(y,1,2) 
     fmt.Println(x)
     fmt.Println(y)
 }
```

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	wg       sync.WaitGroup
	execTime = time.Second
)

func finishReq(timeout time.Duration) int {
	ch := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(execTime)
		ch <-200
	}()
	select {
	case result := <- ch:
		return result
	case <-time.After(timeout):
		return -1
	}
}

func main()  {
	timeout := 50 * time.Microsecond
	fmt.Printf("Result:%d\n",finishReq(timeout))
	wg.Wait()
}

```

---

```go

```


1.slice、map、struct、channel、goroutine
2.iota、new、make、for-range、reflect、defer
3.error、panic、recover