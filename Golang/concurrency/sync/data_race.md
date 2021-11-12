# Detecting Race Conditions With Go

- data race：两个或多个goroutine访问同一个资源，并尝试对该资源进行读写而不考虑其他goroutine。
- race detector：Go1.1引入了竞争检测器，它能够检测并报告它发现的任何竞争条件。

```go
// 只需要在执行测试或编译的时候加上 -race的flag就可以开启数据竞争的检测
$ go test -race mypkg    // to test the package
$ go run -race mysrc.go  // to run the source file
$ go build -race mycmd   // to build the command
$ go install -race mypkg // to install the package
```

>不建议在生产环境 build 的时候开启数据竞争检测，因为这会带来一定的性能损失(一般内存5-10倍，执行时间2-20倍)，当然 必须要 debug 的时候除外。
建议在执行单元测试时始终开启数据竞争的检测。


# 例子1

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
var counter int

func main () {
	for i:=0; i< 2; i++ {
		wg.Add(1)
		go routine()
	}
	wg.Wait()
	fmt.Println("Final counter:",counter)
}

func routine() {
	for i:=0; i< 2; i++ {
		value := counter
		time.Sleep(1 * time.Nanosecond) // 产生goroutine的上下文切换。
		value = value+1
		counter = value
	}
	wg.Done()
}
```

```go
D:\Users\sunyuan\GolandProjects\stdlib-go>go build -race Golang\concurrency\sync\data_race.go
D:\Users\sunyuan\GolandProjects\stdlib-go>data_race.exe
==================
WARNING: DATA RACE
Write at 0x0000003fee90 by goroutine 8:
  main.routine()
      D:/Users/80302367/GolandProjects/stdlib-go/Golang/concurrency/sync/data_race.go:26 +0x77

Previous read at 0x0000003fee90 by goroutine 7:
  main.routine()
      D:/Users/80302367/GolandProjects/stdlib-go/Golang/concurrency/sync/data_race.go:23 +0x4e

Goroutine 8 (running) created at:
  main.main()
      D:/Users/80302367/GolandProjects/stdlib-go/Golang/concurrency/sync/data_race.go:15 +0x6f

Goroutine 7 (running) created at:
  main.main()
      D:/Users/80302367/GolandProjects/stdlib-go/Golang/concurrency/sync/data_race.go:15 +0x6f
==================
Final counter: 2
Found 1 data race(s)
```

最终的输出结果2、3、4都有可能。

# data race 配置
通过设置`GORACE`环境变量，可以达到控制data race的行为。
>GORACE="option1=val1 option2=val2"

可选配置：

 配置 | 默认值 | 说明 |
--- |---| ---|
log_path | stderr| 日志文件的路径，除了文件路径外，还支持stderr、stdout两个特殊值。
exitcode | 66 | 退出码
strip_path_prefix | ""| 从日志中的文件信息里面去除相关的前缀，可以去除本地信息，同时会更加好看。
history_size | 1 | per-goroutine内存访问历史记录为32K*2^history_size，增加该值可以避免出现堆栈还原失败的错误，但另一方面会增加内存的使用。
halt_on_error | 0 | 控制出现第一个数据竞争错误后是否立即退出。
atexit_sleep_ms | 100 | 控制main退出之前sleep的时间


# 经典案例
这些案例都是来自Golang的官方文档[Data Race Detector](https://golang.org/doc/articles/race_detector) ,是初学者很容易犯的错误。

## 例子2 在循环中开启goroutine并引用循环变量
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			fmt.Println(i) // Not the 'i' you are looking for.
			wg.Done()
		}()
	}
	wg.Wait()
}
```
新启的goroutine读取`i`的值，main对`i`赋值，会发生data race。最终输出结果是不可预知的，绝大多数情况下会输出55555(`for循环比goroutine启动快`)。

要修复这个bug,只需要将i作为参数传入，让每个goroutine拿到的都是拷贝值。
```go
func main() {
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(j int) {
			fmt.Println(j) // Good. Read local copy of the loop counter.
			wg.Done()
		}(i)
	}
	wg.Wait()
}
```

## 例子3 不小心共享变量

```go
package main

import "os"

func main() {
	ParallelWrite([]byte("Accidentally shared variable"))
}

// ParallelWrite writes data to file1 and file2, returns the errors.
func ParallelWrite(data []byte) chan error {
	res := make(chan error, 2)
	f1, err := os.Create("file1")
	if err != nil {
		res <- err
	} else {
		go func() {
			// This err is shared with the main goroutine,
			// so the write races with the write below.
			_, err = f1.Write(data)
			res <- err
			f1.Close()
		}()
	}
	f2, err := os.Create("file2") // 这里的err变量并没有被重新声明,而是使用上面声明的err变量。
	if err != nil {
		res <- err
	} else {
		go func() {
			_, err = f2.Write(data)
			res <- err
			f2.Close()
		}()
	}
	return res
}
```

在初始化变量的时,如果在同一个作用域下，如下方代码，这里使用的err其实是同一个变量，只是f1、f2不同，具体可以看
[Redeclaration and reassignment](https://golang.org/doc/effective_go#redeclaration) 和 [短变量声明](https://studygolang.com/articles/17913)

所以,上述代码中的两个goroutine共享了`f1, err := os.Create("file1")`这行代码声明的err变量,存在data race。

要修复这个bug,只需在goroutine声明新变量即可。
```go
_,err := f1.Write(data)
_,err := f2.Write(data)
```

## 例子4 未被保护的全局变量

```go
var service map[string]net.Addr

func RegisterService(name string, addr net.Addr) {
	service[name] = addr
}

func LookupService(name string) net.Addr {
	return service[name]
}
```

多个goroutine并发调用上述代码,`service`变量产生data race,并发读写map是不安全的。

针对全局变量,一般的做法是*加锁*，就上述例子也可以使用sync.Map。

```go
var (
	service   map[string]net.Addr
	serviceMu sync.Mutex
)

func RegisterService(name string, addr net.Addr) {
	serviceMu.Lock()
	defer serviceMu.Unlock()
	service[name] = addr
}

func LookupService(name string) net.Addr {
	serviceMu.Lock()
	defer serviceMu.Unlock()
	return service[name]
}
```

## 例子5 未被保护的原始类型变量

Golang类型系统中定义的原始类型也可能发生data race,比如(bool、int、int64,etc...)。
```go
type Watchdog struct{
    last int64 
}

func (w *Watchdog) KeepAlive() {
	w.last = time.Now().UnixNano() // First conflicting access.
}

func (w *Watchdog) Start() {
	go func() {
		for {
			time.Sleep(time.Second)
			// Second conflicting access.
			if w.last < time.Now().Add(-10*time.Second).UnixNano() {
				fmt.Println("No keepalives for 10 seconds. Dying.")
				os.Exit(1)
			}
		}
	}()
}
```
修复上述bug,通常是使用channel或者mutex。针对一些特殊的数据类型，可以使用`sync/atomic`包来实现原子操作。

```go
type Watchdog struct{ 
    last int64 
}

func (w *Watchdog) KeepAlive() {
	atomic.StoreInt64(&w.last, time.Now().UnixNano())
}

func (w *Watchdog) Start() {
	go func() {
		for {
			time.Sleep(time.Second)
			if atomic.LoadInt64(&w.last) < time.Now().Add(-10*time.Second).UnixNano() {
				fmt.Println("No keepalives for 10 seconds. Dying.")
				os.Exit(1)
			}
		}
	}()
}
```
