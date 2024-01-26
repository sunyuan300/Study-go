# interface
Go语言中的接口是一种数据类型，它拥有两种不同的表现形式：
- 带有一组方法的`interface`。
- 不带任何方法的`interface{}`。

在实现上分别使用[runtime.iface](https://github.com/golang/go/blob/6c64b6db6802818dd9a4789cdd564f19b70b6b4c/src/runtime/runtime2.go#L203)和[runtime.eface](https://github.com/golang/go/blob/6c64b6db6802818dd9a4789cdd564f19b70b6b4c/src/runtime/runtime2.go#L208)来表示。
```go
type iface struct {
    tab  *itab
    data unsafe.Pointer
}
```

```go
type eface struct {
    _type *_type
    data  unsafe.Pointer
}
```

