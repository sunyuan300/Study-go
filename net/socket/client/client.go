package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// 1.与Server端建立socket连接
	conn, err := net.Dial("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatalf("connect server failed,err:%v\n", err)
	}

	// 2.向server端发送数据
	_, err = conn.Write([]byte("hello"))
	if err != nil {
		log.Fatalf("send data failed,err:%v\n", err)
	}

	// 3.从服务端接收数据
	buf := make([]byte,128)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("read data failed,err:%v\n", err)
	}
	fmt.Printf("收到数据:%s",string(buf[:n]))
}