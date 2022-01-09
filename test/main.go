package main

import (
	"fmt"
	_ "net/http/pprof"
	"time"
)

func main() {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Println("Hello scheduler")
	}
}
