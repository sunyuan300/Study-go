# 反射

## 基本构成

`reflect`包含两对非常重要**类型**和**函数**。

类型：

- Type
  ```go
  type Type interface {
	        NumMethod() int
            ...
	        PkgPath() string
	        Kind() Kind
            ...        
	        Field(i int) StructField
	        FieldByIndex(index []int) StructField
	        FieldByName(name string) (StructField, bool)
	        FieldByNameFunc(match func(string) bool) (StructField, bool)
            ...
    }
  ```
- Valve:
    ```go
    type Value struct {
	        typ *rtype
	        ptr unsafe.Pointer
	        flag
    }
  
  // 未对外暴露字段，提供了获取或写入数据的方法
  func (v Value) Addr() Value {...}
  func (v Value) Bool() bool {...}
  func (v Value) Bytes() []byte {...}
  ...
    ```

函数：

- TypeOf():获取类型信息
  ```go
  func TypeOf(i interface{}) Type
  ```
- ValveOf():获取数据信息
  ```go
  func ValueOf(i interface{}) Value
  ```

**结论**：`TypeOf()`和`ValueOf()`函数都是将任意类型转换为`Type`或`Value`类型,然后就可以使用对应的方法操作它们。


## 三大法则

1.能将`任意类型`变量转换为反射对象。

2.能将反射对象转换为`interface{}类型`变量。

3.修改反射对象，其值必须可设置。

### 第一法则

将`interface{}`变量转换为反射对象;`TypeOf()`和`ValueOf()`两个函数是连接它们之间的桥梁。
![](https://img.draveness.me/golang-interface-to-reflection.png)

```go
package main

import (
  "fmt"
  "reflect"
)

func main() {
  user := "alex"
  fmt.Println("TypeOf user:", reflect.TypeOf(user)) // TypeOf user: string
  fmt.Println("ValueOf user:",reflect.ValueOf(user)) // ValueOf user: alex
}
```
将变量转换为反射对象后，就可以调用不同的方法获取相关信息。

`Method()`：获取该类型实现的方法

`Field()`：获取该类型包含的全部字段

## 第二法则
反射对象`Value`可以通过`Interface()`方法将反射对象转换为变量，但是变量类型只能是`interface{}类型`,可以通过类型断言将其还原成最原始的类型。

![](https://img.draveness.me/golang-reflection-to-interface.png)
```go
v := reflect.ValueOf(1) // 反射对象Value类型
v.Interface().(int)     // 转换为interface{}类型，再进行类型断言
```

**总结：** 从反射对象到变量的过程是从变量到反射对象的镜面过程，两个过程都需要经历两次转换：

- 变量——>反射对象：

  1.基本类型————>接口类型

  2.接口类型————>反射类型
- 反射对象——>变量：

  1.反射类型————>接口类型

  2.接口类型————>基本类型

![](https://img.draveness.me/golang-bidirectional-reflection.png)