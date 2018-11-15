package main

import (
	"net"
	"github.com/astaxie/beego"
	"fmt"
	"strings"
	"log"
	"os"
)

const(
	LOG_DIRECTORY = "./test.log"
)

var onlineConns = make(map[string]net.Conn)// 存储客户端ip与连接的映射
var messageQueue = make(chan string, 1000) // 消息队列
var quitChan = make(chan bool)
var logger *log.Logger

// 处理错误
func CheckError(err error){
	if err != nil{
		beego.Error(err)
	}
}


func main(){
	// 开启日志
	logFile, err := os.OpenFile(LOG_DIRECTORY, os.O_RDWR|os.O_CREATE, 0)
	if err != nil{
		fmt.Println("log file create failure!")
		os.Exit(-1)
	}
	defer logFile.Close()
	logger = log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)

	// 先监听接口
	listener,err := net.Listen("tcp", "127.0.0.1:8080")
	defer listener.Close()
	CheckError(err)
	fmt.Println("服务端已启动...")
	logger.Println("服务端已启动")
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
	if len(contents) > 2{
		srcAddr := contents[0] // 源地址
		destAddr := contents[1] // 目的地址
		sendMessage := strings.Join(contents[2:], "#") // 发送信息
		destAddr = strings.Trim(destAddr, " ")

		if conn,ok := onlineConns[destAddr]; ok{
			_,err := conn.Write([]byte(srcAddr + "*"+sendMessage))
			CheckError(err)
		}
	}else {
		srcAddr := contents[0] // 源地址
		cmd := contents[1] //命令
		if strings.ToUpper(cmd) == "LIST"{
			var ips string = ""
			for ip := range onlineConns{
				ips = ips + "|" + ip
			}
			if conn, ok := onlineConns[srcAddr]; ok{
				_, err := conn.Write([]byte("服务器端* 目前连接服务器的ip地址: " +ips))
				if err != nil{
					logger.Println(err)
				}
			}
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
	// 协程退出时，将当前链接从onlineConns删除掉
	defer func(conn net.Conn){
		addr := fmt.Sprintf("%s",conn.RemoteAddr())
		delete(onlineConns, addr)
		conn.Close()
		showCurrentUser()
	}(conn)

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
