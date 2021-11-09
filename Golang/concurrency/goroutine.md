## 最佳实践

- 并发性应该交给调用者,而不是函数提供者。
- 调用者应该管控Goroutine的整个生命周期。
  - When goroutine will terminate?(知道goroutine什么时候退出)
  - How to notify a goroutine to terminate?(如何控制goroutine退出)
    - channel
    - context
- log.Fatal()只用在main.main()或者init()


```go
// goroutine的启动不应该交给函数提供者serverApp()

package main

import (
	"fmt"
	"net/http"
)

func serverApp() {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
			fmt.Fprint(resp,"Hello,QCon!")
		})
		http.ListenAndServe(":80",mux)
	}()
}

func main() {
	serverApp()     // 没人知道你启动了一个goroutine

	select {}
}
```

```go
// goroutine的调用者main,无法感知、控制goroutine的退出。
package main

import (
  "fmt"
  "log"
  "net/http"
)

func main() {
  http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprint(resp, "Hello,QCon!")
  })

  go func() {
    if err := http.ListenAndServe(":8080", nil); err != nil {
      log.Fatal(err)
    }
  }()

  select {}
}
```

```go
/*  
    1.调用者main无法感知、控制goroutine的退出。
    2.如果serveDebug()服务很久以前就停止工作了,在进行错误排查时就很困难。
    3.log.Fatal()会调用os.Exit(),会导致defer无法被调用。
*/

package main

import (
  "fmt"
  "log"
  "net/http"
  _ "net/http/pprof"
)

func serveApp() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprint(resp, "Hello,QCon!")
  })

  if err := http.ListenAndServe(":8080", mux); err != nil {
    log.Fatal(err)
  }
}

func serveDebug()  {
  if err := http.ListenAndServe(":8081",http.DefaultServeMux); err != nil {
  	log.Fatal(err)
  }
}

func main() {
  go serveApp()
  go serveDebug()

  select {}
}
```

### 正确姿势

```go
package main

import (
  "context"
  "fmt"
  "net/http"
)

func main() {
  done := make(chan error, 2)   // 感知goroutine退出
  stop := make(chan struct{})   // 控制goroutine退出

  go func() {
    done <- serveDebug(stop)
  }()

  go func() {
    done <- serveApp(stop)
  }()

  for i := 0; i < cap(done); i++ {
    if err := <-done; err != nil {
      fmt.Printf("error:%v", err)
    }
    close(stop)
  }
}

func serveApp(stop <-chan struct{}) error {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprint(resp, "Hello,QCon!")
  })
  s := http.Server{
    Addr:    ":80",
    Handler: mux,
  }

  go func() {
    <-stop  // wait for stop signal
    s.Shutdown(context.Background())
  }()
  
  return s.ListenAndServe()
}

func serveDebug(stop <-chan struct{}) error {
  s := http.Server{
  	Addr: ":8081",
  	Handler: http.DefaultServeMux,
  }
  
  go func() {
  	<-stop  // wait for stop signal
  	s.Shutdown(context.Background())
  }()
  
  return s.ListenAndServe()
}
```
