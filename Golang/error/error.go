package main

import (
	"errors"
	"fmt"
)

// 标准库的errors.New()为什么返回的是指针？

type ErrorString string

func (e ErrorString) Error() string {
	return string(e)
}

// NewErrorString 未返回指针
func NewErrorString(text string) error {
	return ErrorString(text)
}

type ErrorStruct struct {
	s string
}

// NewErrorStruct 未返回指针,和标准库的errors.New()的唯一区别是,标准库返回的是指针。
func NewErrorStruct(text string) error {
	return ErrorStruct{s: text}
}

func (e ErrorStruct) Error() string {
	return e.s
}

var ErrStringType = NewErrorString("EOF")
var ErrStdType = errors.New("EOF")

var ErrStructType = NewErrorStruct("EOF")

func main() {

	// eg:1
	if ErrStringType == NewErrorString("EOF") { // 值比较,所以相等。
		fmt.Println("String Type Error")
	}

	if ErrStdType == errors.New("EOF") { // 指针比较,两个指针执行的内存地址不同,所以不相等。
		fmt.Println("Std Type Error")
	}
	// Output: Named Type Error

	// eg:2

	if ErrStructType == NewErrorStruct("EOF") { // 值比较,所以相等。
		fmt.Println("Struct Type Error")
	}
	// Output: Struct Type Error

	// eg:3
	fmt.Println(errors.New("EOF") == errors.New("EOF"))
	//Output: false
}
