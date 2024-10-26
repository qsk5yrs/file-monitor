package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"os/exec"
	"sync"
	"time"
)

var taskMutex sync.Mutex
var isTaskRunning bool

func log(format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] "+format+"\n", append([]interface{}{timestamp}, v...)...)
}

func longRunningTask() {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	startTime := time.Now()
	log("Test Task started.")
	defer func() {
		if r := recover(); r != nil {
			log("Task panicked at %v: %v", time.Now(), r)
		}
	}()

	// 模拟任务执行
	time.Sleep(2 * time.Second)

	log("Test Task finished, took %v.", time.Since(startTime))
}

func executeRCloneTask(sourceBucket, destBucket string) {
	taskMutex.Lock()
	isTaskRunning = true

	defer func() {
		isTaskRunning = false
		taskMutex.Unlock()
	}()

	startTime := time.Now()
	log("RClone Sync Task started.")
	defer func() {
		if r := recover(); r != nil {
			log("RClone Sync Task panicked at: %v", r)
		}
	}()

	// 定义 rClone 命令及其参数
	cmd := exec.Command("/usr/local/bin/rclone", "sync", sourceBucket, destBucket, "-v", "--delete-excluded", "--transfers=8", "--checkers=16")

	// 捕获命令的标准输出和标准错误
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	// 执行命令
	err := cmd.Run()

	if err != nil {
		log("RClone Sync Task failed: %v", err)
		if errOut.Len() > 0 {
			lines := errOut.String()
			for _, line := range lines {
				log("Command stderr: %s", line)
			}
		}
	} else {
		if out.Len() > 0 {
			lines := out.String()
			for _, line := range lines {
				log("Command stdout: %s", line)
			}
		}
		log("RClone Sync Task finished, took %v", time.Since(startTime))
	}
}

func main() {
	// 定义命令行参数
	startCmd := flag.Bool("start", false, "Start the cron job")
	sourceBucket := flag.String("source", "", "Source bucket name")
	destBucket := flag.String("dest", "", "Destination bucket name")
	interval := flag.Int("interval", 10, "Interval in seconds for running the task")
	helpCmd := flag.Bool("h", false, "Show help message")

	flag.Parse()

	if *helpCmd || !*startCmd || *sourceBucket == "" || *destBucket == "" {
		fmt.Println("Usage: program [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// 如果指定了 -start 参数，则启动定时任务
	if *startCmd {
		c := cron.New()
		_, err := c.AddFunc(fmt.Sprintf("@every %ds", *interval), func() {
			taskMutex.Lock()
			if isTaskRunning {
				taskMutex.Unlock()
				return
			}
			taskMutex.Unlock()
			go func() {
				//longRunningTask()
				executeRCloneTask(*sourceBucket, *destBucket) // 使用新的goroutine来执行任务
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
