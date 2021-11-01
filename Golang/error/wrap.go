package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func a() error {
	//return errors.New("an error")
	return errors.Errorf("an error in a()")
}

func b() error {
	return errors.WithMessage(a(), "an err in b func")
}

func main() {
	err := b()
	fmt.Printf("%+v\n", err)

	// Output:
	/*
		an error in a()
		main.a
			D:/Go_Code/src/stdlib_learn/Golang/error/wrap.go:11
		main.b
			D:/Go_Code/src/stdlib_learn/Golang/error/wrap.go:15
		main.main
			D:/Go_Code/src/stdlib_learn/Golang/error/wrap.go:19
		runtime.main
			D:/Go/src/runtime/proc.go:204
		runtime.goexit
			D:/Go/src/runtime/asm_amd64.s:1374
		an err in b func
	*/
}
