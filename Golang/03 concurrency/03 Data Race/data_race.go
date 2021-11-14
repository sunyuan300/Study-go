/*// eg:1
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
}*/

/*// eg:2
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			fmt.Println(i) // Not the 'i' you are looking for.
			wg.Done()
		}()
	}
	wg.Wait()
}*/

package main

import "os"

func main() {
	ParallelWrite([]byte("xxx"))
}

// ParallelWrite writes data to file1 and file2, returns the errors.
func ParallelWrite(data []byte) chan error {
	res := make(chan error, 2)
	f1, err := os.Create("file1")
	if err != nil {
		res <- err
	} else {
		go func() {
			// This err is shared with the main goroutine,
			// so the write races with the write below.
			_, err = f1.Write(data)
			res <- err
			f1.Close()
		}()
	}
	f2, err := os.Create("file2") // The second conflicting write to err.
	if err != nil {
		res <- err
	} else {
		go func() {
			_, err = f2.Write(data)
			res <- err
			f2.Close()
		}()
	}
	return res
}