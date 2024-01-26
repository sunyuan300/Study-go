
参考：
https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-array-and-slice/
https://geektutu.com/post/hpg-slice.html
https://www.liwenzhou.com/posts/Go/06_slice/

# 数组：

数组类型定义了长度和元素类型，数组的长度固定且是数组类型的一部分。比如[4]int和[5]int是不同的数组类型。类型[4]int对应在内存中四个连续的整数。

Go的数组是值语义。一个数组变量表示整个数组，而不是指向第一个元素的指针(不像C语言的数组)。当一个数组变量被赋值或者被传递时，实际上会复制整个数组。
数组的创建：
```go
b := [2]string{"Penn","Teller"}
b := [...]string{"Penn","Teller"}
// 两种写法都对应[2]string类型
```

# 切片
切片是基于数组的抽象数据类型，和数组不同的是，切片类型没有固定长度，切片的创建：
```go
letters := []string{"a", "b", "c", "d"}
letters := make([]byte,5,5])
letters := make([]byte,5)    // 忽略容量参数时，容量默认值 == 长度参数

// 调用make时，内部会分配一个数组，然后返回数组对应的切片
```

切片的零值为nil，对应的len和cap都是0。
```go
var str1 []string     // str1 == nil    len(str1) == 0 && cap(str1) == 0
str2 := make([]string,0) // str2 != nil len(str2) == 0 && cap(str2) == 0

// 所以不能通过判断 len ==0 && cap == 0 来判断切片是否为nil
```

# 切片的本质
编译期间的切片是 [cmd/compile/internal/types.Slice](https://github.com/golang/go/blob/3b2a578166bdedd94110698c971ba8990771eb89/src/cmd/compile/internal/types/type.go#L346) 类型，
但是在运行期间切片是[reflect.SliceHeader](https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/reflect/value.go#L1994)类型，**每个切片会指向一个底层数组**。
```go
type SliceHeader struct {
    Data uintptr    // 是一个指向数组某个元素的指针
    Len  int
    Cap  int
}

// 就是将数组封装成一个结构体
```

```go
s2 := make([]byte,5)
```

```go
s2 = s1[2:4]

// 本质就是修改    Data指针值    Len长度    Cap容量
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```



新建的切片s2并不会新分配一段内存空间并复制切片s1相应的元素，而是**共用s1的内存空间**，因此当修改新切片的元素会影响原始切片。

# 追加
Go语言的内建函数`append()`可以为切片**动态添加**元素。 可以一次添加一个元素，可以添加多个元素，也可以添加另一个切片中的元素（后面加…）。

注：通过var声明的nil切片可以在append()函数直接使用，无需初始化。
```go
var s1 []int // s1 == nil
s1 = append(s1,1)      // [1]
s1 = append(s1,2,3,4)  // [1,2,3,4]
s2 := []int{5,6,7}
s1 = append(s1,s2...)  // [1,2,3,4,5,6,7]
```


# 扩容
当切片底层数组的容量不够时，使用append()添加元素会自动进行扩容，此时切片指向的***底层数组会重新分配一块地址连续的内存***，并复制原有的元素。

扩容策略：$GOROOT/src/runtime/slice.go
```go
func growslice(et *_type, old slice, cap int) slice {
    newcap := oldCap
    doublecap := newcap + newcap
    if newLen > doublecap {
        newcap = newLen
    } else {
        const threshold = 256
        if oldCap < threshold {
            newcap = doublecap
        } else {
            for 0 < newcap && newcap < newLen {
                newcap += (newcap + 3*threshold) / 4
            }
            if newcap <= 0 {
                newcap = newLen
            }
        }
    }
// ......
}
```
1. 如果期望容量大于当前容量的两倍就会使用期望容量； 
2. 如果当前切片的长度小于 1024 就会将容量翻倍； 
3. 如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；

# 删除
Go语言中并没有删除切片元素的专用方法，我们可以使用切片本身的特性来删除元素。
```go
a := []int{30, 31, 32, 33, 34, 35, 36, 37}
a = append(a[:2], a[3:]...)

// a = append(a[:index], a[index+1:]...)
```


# 遍历(查询)
支持索引遍历和for range遍历
```go
s := []int{1, 3, 5}

for i := 0; i < len(s); i++ {
    fmt.Println(i, s[i])
}

for index, value := range s {
    fmt.Println(index, value)
}
```


# 赋值拷贝(修改)
切片是引用类型，直接赋值拷贝其实都是指向同一块内存地址，修改元素会相互影响。
```go
s1 := make([]int, 3) // [0 0 0]
s2 := s1             // 将s1直接赋值给s2，s1和s2共用一个底层数组
s2[0] = 100          // 由于共用内存，会影响s1的元素，s1 == s2 == [100,0,0]
```
内建的`copy()`函数可以迅速地将一个切片的数据复制到另外一个切片空间中。
```go
a := []int{1, 2, 3, 4, 5}
c := make([]int, 5, 5)
copy(c, a)     // 使用copy()函数将切片a中的元素复制到切片c
c[0] = 1000    // 不会影响切片a的元素
```


# Slice性能陷阱

## 大量内存得不到释放

在切片的基础上进行切片，不会创建新的底层数组。因为原来的底层数组没有发生变化，内存会一直占用，直到没有变量引用该数组。
因此很可能出现这么一种情况，原切片由大量的元素构成，但是我们在原切片的基础上切片，虽然只使用了很小一段，但**底层数组**在内存中仍然占据了大量空间，**得不到释放**。
比较推荐的做法，_**使用 copy 替代 re-slice**_。

# 面试题
01:
```go
package main

import "fmt"

func main() {
    v := make([]int, 0, 5) // len(v) = 0   = 0   底层数组=[0,0,0,0,0]
    v = append(v, 2, 3, 5) // len(v) = 0+3 = 3   底层数组=[2,3,5,0,0]
    a := append(v, 0, -1)  // len(a) = 3+2 = 5   底层数组=[2,3,5,0,-1]
    fmt.Println(v)
    fmt.Println(a)

    b := append(v, 1) // len(b) = 3+1 = 4   底层数组=[2,3,5,1,-1]
    fmt.Println()
    fmt.Println(v)
    fmt.Println(a)
    fmt.Println(b)

    c := append(v, 6, 7, 8, 9) //len(c)=3+4=7 重新分配底层数组=[2,3,5,6,7,8,9] 
    fmt.Println()
    fmt.Println(v)
    fmt.Println(a)
    fmt.Println(b)
    fmt.Println(c)

    d := append(v, 12) // len(d) = 3+1 =4 底层数组=[2,3,5,12,-1]
    fmt.Println()
    fmt.Println(v)
    fmt.Println(a)
    fmt.Println(b)
    fmt.Println(c)
    fmt.Println(d)
}
```

output:
```shell
[2 3 5]
[2 3 5 0 -1]

[2 3 5]
[2 3 5 1 -1]
[2 3 5 1]

[2 3 5]
[2 3 5 1 -1]
[2 3 5 1]
[2 3 5 6 7 8 9]

[2 3 5]
[2 3 5 12 -1]
[2 3 5 12]
[2 3 5 6 7 8 9]
[2 3 5 12]
```
[解析](https://studygolang.com/articles/28949)


02:
```go
package main

import "fmt"

func sliceAppend(slice []int, v int) []int {
    return append(slice, v)
}

func main() {
    arr := make([]int, 5, 10) // len(arr) = 5 底层数组=[0,0,0,0,0,0,0,0,0,0]
    for i := 0; i < 5; i++ {
        arr = append(arr, i)
    }
    // len(arr) = 5+5 = 10 底层数组= [0,0,0,0,0,0,1,2,3,4]

    slice := arr[:] // len(slice) = 10  底层数组= [0,0,0,0,0,0,1,2,3,4]

	//len(slice2) = 10+1 = 11  重新分配底层数组= [0,0,0,0,0,0,1,2,3,4,1000]
    slice2 := sliceAppend(slice, 1000)

    slice3 := arr[:2]
    slice3 = append(slice3, 10) //len(slice3)=2+1=3 底层数组=[0,0,10,0,0,0,1,2,3,4]

    slice4 := arr[:8]
    sliceAppend(slice4, 100) //len(slice4)=8 底层数组=[0,0,10,0,0,0,1,2,100,4]

    fmt.Println(arr, len(arr), cap(arr))
    fmt.Println(slice, len(slice), cap(slice))
    fmt.Println(slice2, len(slice2), cap(slice2))
    fmt.Println(slice3, len(slice3), cap(slice3))
    fmt.Println(slice4, len(slice4), cap(slice4))
}
```

```shell
output:
[0 0 10 0 0 0 1 2 100 4] 10 10
[0 0 10 0 0 0 1 2 100 4] 10 10
[0 0 0 0 0 0 1 2 3 4 1000] 11 20
[0 0 10] 3 10
[0 0 10 0 0 0 1 2] 8 10
```


03:
// 下面两段代码输出什么
```go
func main() {
    s := make([]int, 5)    //len(s)=5     底层数组=[0,0,0,0,0] 
    s = append(s, 1, 2, 3) //len(s)=5+3=8 重新分配底层数组=[0,0,0,0,0,1,2,3]
    fmt.Println(s)         // [0 0 0 0 0 1 2 3]
}

func main() {
    x := make([]int, 0)       // len(x)=0      底层数组=[]
    y = append(x, 1, 2, 3, 4) // len(y)=0+4     重新分配底层数组=[1,2,3,4]
    fmt.Println(x)            // []
    fmt.Println(y)            // [1,2,3,4]
}
```



04:
```go
//是否可以通过编译？

func main() {
    list := new([]int)
    list = append(list, 1)
    fmt.Println(list)
}

// 不能,因为new返回一个指针,append函数是对切片的操作,而不是指针。
```


05：
```go
// 是否可以通过编译？
package main

import "fmt"

func main() {
    s1 := []int{1, 2, 3}
    s2 := []int{4, 5}
    s1 = append(s1, s2)
    fmt.Println(s1)
}

// 不能,正确应该：s1 = append(s1, s2...)
```

