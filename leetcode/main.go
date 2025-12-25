package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Counter struct {
	incCh   chan struct{}
	readCh  chan chan int
	closeCh chan struct{}
}

func (c *Counter) Inc() {
	select {
	// 发送自增信号
	case c.incCh <- struct{}{}:
	// 监听关闭的信号
	case <-c.closeCh:
		fmt.Println("Counter closed")
	}
}
func (c *Counter) Read() int {
	resp := make(chan int)
	select {
	case c.readCh <- resp:
		return <-resp
	case <-c.closeCh:
		fmt.Println("Counter closed")
		return -1
	}
}

func (c *Counter) Close() {
	close(c.closeCh)
}

func NewCounter() *Counter {
	counter := &Counter{
		incCh:   make(chan struct{}),
		readCh:  make(chan chan int),
		closeCh: make(chan struct{}),
	}
	go func() {
		val := 0
		for {
			select {
			// 监听自增信号，执行自增
			case <-counter.incCh:
				val++
			// 监听读取信号，写入接收者的 chan
			case resp := <-counter.readCh:
				resp <- val
			case <-counter.closeCh:
				fmt.Println("Counter closed")
				return
			}

		}
	}()
	return counter
}

func main() {
	// 1. 打印初始 Goroutine 数量
	printGoroutineNum("程序启动时")

	// 创建实例
	csp := NewCounter()

	// 等待一下让后台协程跑起来
	time.Sleep(100 * time.Millisecond)
	printGoroutineNum("创建 CSP Counter 后")

	// 使用并发写入
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			csp.Inc()
		}()
	}
	wg.Wait()

	fmt.Printf("计数结果: %d\n", csp.Read())

	// 关键步骤：关闭计数器
	fmt.Println("正在关闭 CSP Counter...")
	csp.Close()

	// 等待 runtime 回收 Goroutine
	time.Sleep(200 * time.Millisecond)
	printGoroutineNum("关闭 CSP Counter 后") // 这里应该恢复到和启动时一样（或接近）

	fmt.Println("\n程序结束")
}

// 辅助函数：打印当前正在运行的 Goroutine 数量
func printGoroutineNum(tag string) {
	num := runtime.NumGoroutine()
	fmt.Printf("当前 Goroutine 数量 [%s]: %d\n", tag, num)
}
