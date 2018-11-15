package main

import (
	"net"
	"github.com/astaxie/beego"
	"fmt"
)

// 处理错误
func CheckError(err error){
	if err != nil{
		beego.Error(err)
	}
}

func main(){
	// 先监听接口
	listener,err := net.Listen("tcp", "127.0.0.1:8080")
	CheckError(err)
	fmt.Println("服务端已启动...")
	// 然后针对每一个请求，开启一个协程进行处理
	for {
		conn,err := listener.Accept()
		CheckError(err)
		go ProcessInfo(conn)
	}
}

// 处理客户端的请求
func ProcessInfo(conn net.Conn){
	defer conn.Close()
	// 申请缓存
	buf := make([]byte, 1024)
	for{
		num,_ :=conn.Read(buf)
		if num >0 {
			fmt.Println(string(buf))
		}
	}
}
