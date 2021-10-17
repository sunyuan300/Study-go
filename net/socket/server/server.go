package main

import (
	"fmt"
	"log"
	"net"
)

const Address = "127.0.0.1:9090"

func main() {
	// 1.监听本地的9090端口
	listener, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalf("faild to listen:%v", err)
	}
	log.Printf("listen to %s\n",Address)

	for {
		// 2.接收来自客户端的请求连接
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept failed,err:%v", err)
			continue
		}
		// 3.将请求连接交给一个goroutine处理
		go process(conn)
	}
}

func process(conn net.Conn) {
	// 1.处理完成后关闭连接
	defer conn.Close()
	for {
		// 2.读取请求连接数据
		buf := make([]byte,128)
		n,err := conn.Read(buf)
		if err != nil {
			log.Printf("read from conn failed,err:%v\n", err)
			break
		}
		fmt.Printf("收到数据:%s\n",string(buf[:n]))

		// 3.向请求连接发送数据
		_,err = conn.Write([]byte("ok"))
		if err != nil {
			log.Printf("write data to conn failed,err:%v",err)
			break
		}
	}
}