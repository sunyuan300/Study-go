# Detecting Race Conditions With Go

- data race：两个或多个goroutine访问同一个资源，并尝试对该资源进行读写而不考虑其他goroutine。
- race detector：Go1.1引入了竞争检测器，它能够检测并报告它发现的任何竞争条件。

```go
// 只需要在执行测试或编译的时候加上 -race的flag就可以开启数据竞争的检测
go build -race xx.go
go test -race xx.go
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