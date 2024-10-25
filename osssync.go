package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

var taskMutex sync.Mutex

func log(format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] "+format+"\n", append([]interface{}{timestamp}, v...)...)
}

func longRunningTask() {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	startTime := time.Now()
	log("Task started.")
	defer func() {
		if r := recover(); r != nil {
			log("Task panicked at %v: %v", time.Now(), r)
		}
	}()

	// 模拟任务执行
	time.Sleep(15 * time.Second)

	log("Task finished, took %v.", time.Since(startTime))
}

func main() {
	// 定义命令行参数
	startCmd := flag.Bool("start", false, "Start the cron job")
	helpCmd := flag.Bool("h", false, "Show help message")

	flag.Parse()

	if *helpCmd || !*startCmd && flag.NArg() == 0 {
		fmt.Println("Usage: program [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// 如果指定了 -start 参数，则启动定时任务
	if *startCmd {
		c := cron.New()
		_, err := c.AddFunc("@every 10s", func() {
			go func() {
				longRunningTask() // 使用新的goroutine来执行任务
			}()
		})
		if err != nil {
			log("Error adding cron job: %v", err)
			return
		}
		c.Start()

		// 阻塞主goroutine，防止程序退出
		select {}
	}
}
