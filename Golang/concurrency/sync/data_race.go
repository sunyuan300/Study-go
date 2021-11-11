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