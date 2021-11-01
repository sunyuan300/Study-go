# 最佳实践

## panic

1. 在程序启动的时候，如果有**强依赖**的服务出现故障时 panic 退出。
2. 在程序启动的时候，如果发现有配置明显不符合要求，可以 panic 退出（防御编程）。
3. 其他情况下只要不是不可恢复的程序错误，都不应该直接 panic 应该返回 error。
4. 在程序入口处，例如 gin 中间件需要使用 recover 预防 panic 程序退出。
5. 在程序中我们应该避免使用野生的 goroutine：
   1. 如果是在请求中需要执行异步任务，应该使用异步 `worker` ，消息通知的方式进行处理，避免请求量大时大量 goroutine 创建。
   2. 如果需要使用 goroutine，应该使用`统一的Go函数`进行创建，这个函数中会进行 recover ，避免因为野生 goroutine panic 导致主进程退出。

```go
func Go(f func()){
    go func(){
        defer func(){
            if err := recover(); err != nil {
                log.Printf("panic: %+v", err)
            }
        }()
        f()
    }()
}
```

## error

1. 我们在应用程序中使用`github.com/pkg/errors`处理应用错误，注意在公共库当中，我们一般不使用这个。
2. `error`应该是函数的最后一个返回值，当 error 不为 nil 时，函数的其他返回值是不可用的状态，不应该对其他返回值做任何期待。
   1. `func f() (io.Reader, *S1, error)`在这里，我们不知道 io.Reader 中是否有数据，可能有，也有可能有一部分。
3. 错误处理的时候应该先判断错误，`if err != nil` 出现错误及时返回，使代码是一条流畅的直线，避免过多的嵌套。
```go
// good case
func f() error {
    a, err := A()
    if err != nil {
        return err
    }

    // ... 其他逻辑
    return nil
}

// bad case
func f() error {
    a, err := A()
    if err == nil {
    // 其他逻辑
    }

    return err
}
```
4. 在**应用程序中**出现错误时，使用 errors.New 或者 errors.Errorf 返回错误。
```go
func (u *usecese) usecase1() error {
    money := u.repo.getMoney(uid)
    if money < 10 {
        errors.Errorf("用户余额不足, uid: %d, money: %d", uid, money)
    }
    // 其他逻辑
    return nil
}
```
5. 如果是**调用应用程序的其他函数**出现错误，请直接返回，如果需要携带信息，请使用 `errors.WithMessage`。
```go
func (u *usecese) usecase2() error {
    name, err := u.repo.getUserName(uid)
    if err != nil {
        return errors.WithMessage(err, "其他附加信息")
    }

    // 其他逻辑
    return nil
}
```
6. 如果是调用其他库（标准库、企业公共库、开源第三方库等）获取到错误时，请使用 errors.Wrap 添加堆栈信息。
   1. 切记，不要每个地方都是用 errors.Wrap 只需要在错误***第一次***出现时进行 errors.Wrap 即可。
   2. 根据场景进行判断是否需要将其他库的原始错误吞掉，例如可以把 `repository` 层的数据库相关错误吞掉，返回业务错误码，避免后续我们分割微服务或者更换 `ORM` 库时需要去修改上层代码。
   3. 注意我们在基础库，被大量引入的第三方库编写时一般不使用 errors.Wrap 避免堆栈信息重复。
```go
func f() error {
    err := json.Unmashal(&a, data)
    if err != nil {
        return errors.Wrap(err, "其他附加信息")
    }

    // 其他逻辑
    return nil
}
```
7. **禁止**在每个出错的地方都打日志，**只需**在进程最开始的地方使用`%+v`进行统一打印，例如http/rpc服务的中间件。
8. 错误判断，使用`errors.Is`进行比较。
```go
func f() err {
	err := A()
	if errors.Is(err,io.EOF) {
	    return nil	
    }
	// 其他业务逻辑
	return nil
}
```
9. 错误类型判断，使用`errors.As`进行赋值。
```go
func f() error {
	err := A()
	var errA errorA
	if errors.As(err,&errA) {
	    // ...	
    }
	return nil
}
```
10. 对于业务错误，推荐在统一的地方创建一个map,map应该包含错误的code,并在日志中作为独立字段打印,方便做业务告警,这类错误必须有清晰的错误文档。
11. 对于不需要返回、被忽略的错误，必须输出到日志。
12. 对同一类型的错误，采用相同的模式，例如参数错误，不要有的返回404、有的返回200。
13. 处理错误的时候，需要处理已分配的资源，使用`defer`进行清理，例如文件句柄。

## panic or error?

1. 在Go中panic会导致程序直接退出，是一个致命的错误，如果使用`recover`进行处理的话，会存在很多问题：
   1. 性能问题，频繁`panic、recover`性能不好。
   2. 容易导致程序异常退出，只要有一个地方没有处理到就会导致程序退出。
   3. 不可控，一旦`panic`就将处理逻辑移交给了外部，并不能预设外部一定会进行处理。
2. 什么时候使用panic？
   1. 不可恢复的程序错误，例如：索引越界、不可恢复的环境问题、栈溢出，我们才使用panic。
3. 使用error处理有哪些好处？
   1. 简单。
   2. 考虑失败，而不是成功(*Plan for failure,not success*)。
   3. 没有隐藏的控制流。
   4. 完全交给你来控制error。
   5. Error are values。

## 错误定义与判断

### Sentinel Error(预定义错误)
就是定义一些包级别的错误变量，然后在调用的时候外部包可以直接对比变量进行判定，在标准库当中大量的使用了这种方式，例如`io`库中定义的错误。

这种错误处理方式是，将`错误的值`暴露给了外部，这样会导致在做重构或者升级的时候很麻烦，并且这种方式包含的错误信息会十分的有限。

结论：需要提供给他人使用的pkg,尽量避免Sentinel Error；针对业务代码，可以使用Sentinel Error。
```go
// EOF is the error returned by Read when no more input is available.
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either ErrUnexpectedEOF or some other error
// giving more detail.
var EOF = errors.New("EOF")

var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNoProgress = errors.New("multiple Read calls return no data or error")
```
在外部判定的时候一般使用`等值`判断或使用`errors.Is`进行判断。
```go
if err == io.EOF {
	//...
}

if errors.Is(err, io.EOF){
	//...
}
```

### error types(自定义错误)
自定义数据类型并实现`error`接口，然后在外部通过**类型断言**来判断错误类型。这种方式相对哨兵模式，可以包含更加丰富的信息，但同样也是将`错误的类型`暴露给了外部，`这种模型会导致和调用者产生强耦合，从而导致API变得脆弱`，例如标准库中的`os.PathError`。

结论:避免错误类型，或者至少避免将它们作为公共API的一部分。
```go
type MyStruct struct {
	s string
	name string
	path string
}



// 使用的时候
func f() {
    switch err.(type) {
        case nil:
        	//
        case *MyStruct:
            // ...
        case others:
            // ...
    }
}
```

### Opaque errors
这种方式最大的特点是只返回错误，但不暴露错误类型，错误的类型判断通过包暴露的API接进行判断。这种方式我们可以断言错误`实现了特定的行为`，而不是断言错误是特定的类型或值。同时可以减少API的暴露，后续处理比较灵活，这个一般用在公共库比较好。

```go
type temporary interface {
	Temporary() bool
}

func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}
```

## 错误处理的优化


## 错误包装

查看`pkg/errors`源码,可以发现`Wrap`方法除了使用`WithMessage`附加错误信息外,还使用`WithStack`添加了堆栈信息,这样打印错误日志时就可以打印堆栈信息了。
```go
// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	err = &withMessage{
		cause: err,
		msg:   message,
	}
	return &withStack{
		err,
		callers(),
	}
}
```

例子1：
```go
func main() {
	fmt.Printf("err: %+v", c())
}

func a() error {
	return errors.Wrap(fmt.Errorf("xxx"), "test")
}

func b() error {
	return a()
}

func c() error {
	return b()
}
```
输出：
```go
err: xxx
test
main.a
        /home/ll/project/Go-000/Week02/blog/wrap.go:14
main.b
        /home/ll/project/Go-000/Week02/blog/wrap.go:18
main.c
        /home/ll/project/Go-000/Week02/blog/wrap.go:22
main.main
        /home/ll/project/Go-000/Week02/blog/wrap.go:10
runtime.main
        /usr/local/go/src/runtime/proc.go:204
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1374
```

## Wrap最佳实践(pkg/errors)
- 在应用代码(非基础库),使用`errors.New`或者`errors.Errorf`返回错误。`errors.New`和`errors.Errorf`都会保留堆栈信息。
```go
func parseArgs(args []string) error {
	if len(args) < 3 {
		return errors.Errorf("not enough arguments,expected at least 3")
    }
    //...
}
```
- 如果是调用项目中的函数,通常直接简单的返回。
```go
if err != nil {
	return err
}
```
- 如果和第三方/标准库进行协作,考虑使用`errors.Wrap`或者`errors.Wrapf`保存堆栈信息。直白讲就是最底层错误需要包裹,比如和数据库、rpc、第三方库交互的时候。
```go
f,err := os.Open(path)
if err != nil {
	return errors.Wrapf(err,"failed to open %q",err)
}
```
- 直接返回错误,而不是每个错误产生的地方都打印日志。
- 在程序顶层或者工作goroutine顶部(请求入口),使用`%+v`打印详细的堆栈信息。
- 使用`errors.Cause`获取root error(最底层的error),再和sentinel error进行判断。

## 标准库errors.Is、errors.As

### errors.Is
判断err链中是否存在与目标错误匹配的错误类型。
```go
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}
	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func Unwrap(err error) error {
	u, ok := err.(interface {Unwrap() error})
	if !ok {
		return nil
	}
	return u.Unwrap()
}
```

### errors.As
将err转换为指定错误类型(前提是err链中包含与目标错误类型相匹配错误类型)