package main

import (
	"net"
	"github.com/astaxie/beego"
	"fmt"
	"strings"
)

var onlineConns = make(map[string]net.Conn)// 存储客户端ip与连接的映射
var messageQueue = make(chan string, 1000) // 消息队列
var quitChan = make(chan bool)

// 处理错误
func CheckError(err error){
	if err != nil{
		beego.Error(err)
	}
}


func main(){
	// 先监听接口
	listener,err := net.Listen("tcp", "127.0.0.1:8080")
	defer listener.Close()
	CheckError(err)
	fmt.Println("服务端已启动...")

	go ConsumeMessage()

	// 然后针对每一个请求，开启一个协程进行处理
	for {
		conn,err := listener.Accept()
		CheckError(err)

		// 将conn存储到onlineConns映射表中
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		onlineConns[addr] = conn
		showCurrentUser()

		go ProcessInfo(conn)
	}
}


// 专门处理信息的协程
func ConsumeMessage(){
	for{
		select {
		case message := <-messageQueue:
			doProcessMessage(message)
		case <-quitChan:
			break
		}
	}
}

// 转发信息
func doProcessMessage(message string){
	contents := strings.Split(message, "#")
	if len(contents) > 1{
		srcAddr := contents[0] // 源地址
		destAddr := contents[1] // 目的地址
		sendMessage := contents[2] // 发送信息
		destAddr = strings.Trim(destAddr, " ")

		if conn,ok := onlineConns[destAddr]; ok{
			_,err := conn.Write([]byte(srcAddr + "*"+sendMessage))
			CheckError(err)
		}
	}
}

// 显示目前连接上服务器端的机器
func showCurrentUser(){
	fmt.Println("目前连接上服务器端的机器：")
	for i := range onlineConns{
		fmt.Println(i)
	}
}

// 处理客户端的请求
func ProcessInfo(conn net.Conn){
	defer conn.Close()

	buf := make([]byte, 1024)
	remoteAddr := conn.RemoteAddr()
	remoteAddrStr := fmt.Sprintf("%s",remoteAddr)
	for{
		num,_ :=conn.Read(buf)
		if num >0 {
			// 将消息放入到消息队列
			messageQueue <- remoteAddrStr + "#" + string(buf[0:num])
			//fmt.Printf("客户端(%s): %s\n", remoteAddrStr, string(buf[0:num]))
		}
	}
}
