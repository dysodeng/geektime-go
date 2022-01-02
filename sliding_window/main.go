package main

import (
	"container/ring"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	limitCount  int        = 10 // 6s限频
	limitBucket int        = 6  // 滑动窗口个数
	curCount    int32      = 0  // 记录限频数量
	head        *ring.Ring      // 链表
)

func handle(conn *net.Conn) {
	defer (*conn).Close()
	n := atomic.AddInt32(&curCount, 1)
	if n > int32(limitCount) {
		atomic.AddInt32(&curCount, -1)
		(*conn).Write([]byte("HTTP/1.1 404 NOT FOUND\r\n\r\nError, too many request, please try again."))
	} else {
		mu := sync.Mutex{}
		mu.Lock()
		pos := head.Prev()
		val := pos.Value.(int)
		val++
		pos.Value = val
		mu.Unlock()
		time.Sleep(1 * time.Second)
		(*conn).Write([]byte("HTTP/1.1 200 OK\r\n\r\nI can change the world!"))
	}
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:9090") //获取一个tcpAddr
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr) //监听一个端口
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	// 初始化滑动窗口
	head = ring.New(limitBucket)
	for i := 0; i < limitBucket; i++ {
		head.Value = 0
		head = head.Next()
	}
	// 启动执行器
	go func() {
		timer := time.NewTicker(time.Second * 1)
		for range timer.C { // 定时每隔1秒刷新一次滑动窗口数据
			subCount := int32(0 - head.Value.(int))
			newCount := atomic.AddInt32(&curCount, subCount)

			arr := [6]int{}
			for i := 0; i < limitBucket; i++ {
				arr[i] = head.Value.(int)
				head = head.Next()
			}
			fmt.Println("move subCount,newCount,arr", subCount, newCount, arr)
			head.Value = 0
			head = head.Next()
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handle(&conn)
	}
}
