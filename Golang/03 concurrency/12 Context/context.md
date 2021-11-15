# 引入
现在有一个 Server 服务在执行，当请求来的时候我们启动一个 goroutine 去处理，然后在这个 goroutine 当中有对下游服务的 rpc 调用，也会去请求数据库获取一些数据，这时候如果下游依赖的服务比较慢，但是又没挂，只是很慢，可能一次调用要 1min 才能返回结果，这个时候我们该如何处理？

## 场景1
所有下游服务都能正常访问：

使用WaitGroup进行控制，等待所有的goroutine处理完成后返回，但实际耗时远大于用户可容忍时间。
![](wait_group.svg)

## 场景2
部分下游服务故障：

比如rpc goroutine很早就报错了，但是db goroutine还在继续执行，最后程序返回了错误信息，很明显db goroutine执行的这段时间是在浪费用户的时间和系统的资源。
![](no_context.svg)

通过引入Context就可以解决上述问题。

场景1中引入context之后，当上层req goroutine的context超时之后，就会将取消信号同步到下面的所有goroutine，达到超时控制的作用。
![](context_timeout.svg)

场景2中引入context之后，当rpc调用失败后，会发出context取消，并同步到其他goroutine中。
![](context_cancel.svg)

# Context
在实际生活中，一个网络请求都需要开启一个goroutine进行处理，在高并发的情况下(比如618、双11...)，需要成千上万个goroutine，同时每个goroutine可能会开启新的goroutine。在这种场景下，Go提供了Context来跟踪goroutine，从而实现对它们的控制。

## 单goroutine
```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("监控退出，停止了...")
				return
			default:
				fmt.Println("goroutine监控中...")
				time.Sleep(2 * time.Second)
			}
		}
	}(ctx)

	time.Sleep(10 * time.Second)
	fmt.Println("可以了，通知监控停止")
	cancel()
}
```

## 多goroutine
```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go watch(ctx,"【监控1】")
	go watch(ctx,"【监控2】")
	go watch(ctx,"【监控3】")

	time.Sleep(10 * time.Second)
	fmt.Println("可以了，通知监控停止")
	cancel()
}

func watch(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println(name,"监控退出，停止了...")
			return
		default:
			fmt.Println(name,"goroutine监控中...")
			time.Sleep(2 * time.Second)
		}
	}
}
```
context.Background()返回一个空Context实例，一般用于整个Context树的根节点。context.WithCancel(parent)返回一个可取消的子Context和CancelFunc类型的取消函数。所以基于这个Context的子Context都会被取消函数控制，从而优雅的控制goroutine的结束。

# Context源码分析

## 1.接口定义
```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)

	Done() <-chan struct{}

	Err() error

	Value(key interface{}) interface{}
}
```
- Deadline()：deadline表示截至时间，Context会自动在这个时间点发起取消请求；ok表示是否有设置截至时间。
- Done()：返回一个只读struct{}类型的通道，通过判断chan是否可以读取，来判断parent context是否已发起了取消请求。
- Err()：返回Context被取消的原因。
- Value()：获取该Context上绑定的值。

在未调用取消函数前，Done()方法会返回一个未关闭的只读通道，从中取不到任何值；当调用了取消函数后，Done()方法将返回一个已关闭的只读通道，此时可以从中读到通道类型的默认零值，这意味着Context收到了取消信号。

## 2.接口实现

### a.官方实现
```go
type emptyCtx int

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func Background() Context {
	return background
}

func TODO() Context {
	return todo
}

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}
```

Go内置实现了2个，通常以这两个内置作为最顶层的partent context，衍生出更多的子Context。底层类型都是emptyCtx，是一个不可取消，没有设置截止时间，没有携带任何值的Context。
- Background：主要用于main函数、初始化和测试代码中。
- TODO：未定义具体的应用场景，在不知道该使用什么Context的时候，可以使用这个。

### b.第三方实现
```go
// Gin框架中的Context
type Context struct {
   writermem responseWriter
   Request   *http.Request
   Writer    ResponseWriter

   Params   Params
   handlers HandlersChain
   index    int8
   fullPath string

   engine *Engine
   params *Params

   mu sync.RWMutex
   Keys map[string]interface{}
   Errors errorMsgs
   Accepted []string
   queryCache url.Values
   formCache url.Values
   sameSite http.SameSite
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
   return
}

func (c *Context) Done() <-chan struct{} {
   return nil
}

func (c *Context) Err() error {
   return nil
}

func (c *Context) Value(key interface{}) interface{} {
   if key == 0 {
      return c.Request
   }
   if keyAsString, ok := key.(string); ok {
      val, _ := c.Get(keyAsString)
      return val
   }
   return nil
}
```
Gin框架对Context的实现很简单，首先定义一个结构体类型，然后实现Context接口对应的方法即可。

### c.自定义实现
1. 自定义类型(自定义结构体、嵌套结构体、Type MyContext string...)
2. 实现Context接口中的方法
```go
// 基于Gin框架中的Context添加一些需要的字段
type RequestContext struct {
   Context *gin.Context
   User string
   Time time.Time
}

func (ctx *RequestContext) RequestUser() string {
   return ctx.User
}

func (ctx *RequestContext) RequestTime() time.Time {
   return ctx.Time
}


func (ctx *RequestContext) Deadline() (deadline time.Time, ok bool) {
   return
}


func (ctx *RequestContext) Done() <-chan struct{} {
   return nil
}

func (ctx *RequestContext) Err() error {
   return nil
}

func (ctx *RequestContext) Value(interface{}) interface{} {
   return nil
}
```
## 3.Context的继承衍生
```go
type CancelFunc func()

func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithValue(parent Context, key, val interface{}) Context
```

这四个`With`函数，接收的都有一个partent参数，即父Context，**基于这个父Context创建出子Context**的意思。
- WithDeadline()和WithTimeout()的区别：
 - WithDeadline()：到deadline这个时间点自动取消Context，比如22:02:33、08:52:35....
 - WithTimeout()：经过timeout后自动取消Context，比如5s、1min、1hour.....
- WithValue()：与取消Context无关，它是为了生成一个绑定了键值对数据的子Context，即这个子Context带有对应键值对。这里的key必须是可比较的，应该为自定义类型，而不应该是内置类型，以避免使用上下文在包之间发生冲突；同时Value值要是线程安全的。

## 4.WithValue传递元数据
```go
type TraceCode string

func main() {
    key := TraceCode("TRACE_CODE")
	ctx, cancel := context.WithCancel(context.Background())
	//附加值
	valueCtx:=context.WithValue(ctx,key,"【监控1】")
	go watch(valueCtx)
	time.Sleep(10 * time.Second)
	fmt.Println("可以了，通知监控停止")
	cancel()
}

func watch(ctx context.Context) {
    key := TraceCode("TRACE_CODE")
	for {
		select {
		case <-ctx.Done():
			//取出值
			fmt.Println(ctx.Value(key),"监控退出，停止了...")
			return
		default:
			//取出值
			fmt.Println(ctx.Value(key),"goroutine监控中...")
			time.Sleep(2 * time.Second)
		}
	}
}
```

# 使用场景
## 超时控制
```go
ackage main

import (
	"context"
	"fmt"
	"time"
)

// 模拟一个耗时的操作
func rpc() (string, error) {
	time.Sleep(100 * time.Millisecond)
	return "rpc done", nil
}

type result struct {
	data string
	err  error
}

func handle(ctx context.Context, ms int) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
	defer cancel()

	r := make(chan result)
	go func() {
		data, err := rpc()
		r <- result{data: data, err: err}
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("timeout: %d ms, context exit: %+v\n", ms, ctx.Err())
	case res := <-r:
		fmt.Printf("result: %s, err: %+v\n", res.data, res.err)
	}
}

func main() {
	// 这里模拟接受请求，启动一个协程去发起请求
	for i := 1; i < 5; i++ {
		time.Sleep(1 * time.Second)
		go handle(context.Background(), i*50)
	}

	// for test, hang
	time.Sleep(time.Second)
}

/*
Output:
    timeout: 50 ms, context exit: context deadline exceeded
    result: rpc done, err: <nil>
    result: rpc done, err: <nil>
    result: rpc done, err: <nil>
*/
```
## 错误取消
```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func f1(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("f1: %w", ctx.Err())
	case <-time.After(time.Millisecond): // 模拟短时间报错
		return fmt.Errorf("f1 err in 1ms")
	}
}

func f2(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("f2: %w", ctx.Err())
	case <-time.After(time.Hour): // 模拟一个耗时操作
		return nil
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := f1(ctx); err != nil {
			fmt.Println(err)
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		if err := f2(ctx); err != nil {
			fmt.Println(err)
			cancel()
		}
	}()

	wg.Wait()
}
/*
Output:
    f1 err in 1ms
    f2: context canceled
*/
```
## 跨goroutine数据同步
一般会用来传递 tracing id, request id 这种数据。
```go
const requestIDKey int = 0

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			// 从 header 中提取 request-id
			reqID := req.Header.Get("X-Request-ID")
			// 创建 valueCtx。使用自定义的类型，不容易冲突
			ctx := context.WithValue(
				req.Context(), requestIDKey, reqID)

			// 创建新的请求
			req = req.WithContext(ctx)

			// 调用 HTTP 处理函数
			next.ServeHTTP(rw, req)
		}
	)
}

// 获取 request-id
func GetRequestID(ctx context.Context) string {
	ctx.Value(requestIDKey).(string)
}

func Handle(rw http.ResponseWriter, req *http.Request) {
	// 拿到 reqId，后面可以记录日志等等
	reqID := GetRequestID(req.Context())
	...
}

func main() {
	handler := WithRequestID(http.HandlerFunc(Handle))
	http.ListenAndServe("/", handler)
}
```
## 防止goroutine泄露
gen 这个函数中如果不使用 context done 来控制的话就会导致 goroutine 泄漏，因为这里面的 for 是一个死循环，没有 ctx 就没有相关的退出机制。
```go
func main() {
	// gen generates integers in a separate goroutine and
	// sends them to the returned channel.
	// The callers of gen need to cancel the context once
	// they are done consuming generated integers not to leak
	// the internal goroutine started by gen.
	gen := func(ctx context.Context) <-chan int {
		dst := make(chan int)
		n := 1
		go func() {
			for {
				select {
				case <-ctx.Done():
					return // returning not to leak the goroutine
				case dst <- n:
					n++
				}
			}
		}()
		return dst
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers

	for n := range gen(ctx) {
		fmt.Println(n)
		if n == 5 {
			break
		}
	}
}
```