package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var fileEventTimes = make(map[string]time.Time) // 记录每个文件的最后处理时间
// var fileSizes = make(map[string]int64)          // 记录文件大小
var debounceDuration1 = 1 * time.Second
var debounceDuration = 10 * time.Second    // 静默期，确保文件没有新的写入操作
var checkInterval = 500 * time.Millisecond // 文件大小检测间隔

func main() {
	watchDir := "d:/testDir" // 监控的目录
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					fileInfo, err := os.Stat(event.Name)
					if err == nil && fileInfo.IsDir() {
						log.Println("New directory detected:", event.Name)
						addWatcherRecursive(watcher, event.Name)
					} else {
						mu.Lock()
						if shouldProcess(event.Name) {
							wg.Add(1)
							go checkIfFileComplete(event.Name, &wg) // 检测文件是否写入完成
						}
						mu.Unlock()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	addWatcherRecursive(watcher, watchDir)

	<-make(chan struct{})
}

// 递归添加目录及其子目录的监听
func addWatcherRecursive(watcher *fsnotify.Watcher, dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			log.Println("Watching directory:", path)
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error watching directories:", err)
	}
}

// 判断是否应该处理该文件（基于时间窗口去重）
func shouldProcess(filePath string) bool {
	lastProcessed, exists := fileEventTimes[filePath]
	if !exists || time.Since(lastProcessed) > debounceDuration1 {
		fileEventTimes[filePath] = time.Now()
		return true
	}
	return false
}

// 检测文件是否已经写入完成
func checkIfFileComplete(filePath string, wg *sync.WaitGroup) {
	defer wg.Done()

	var prevSize int64 = -1
	var stableDuration time.Duration

	for {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Println("Error reading file info:", err)
			return
		}

		currentSize := fileInfo.Size()

		// 如果文件大小在 checkInterval 时间内没有变化，则认为文件稳定
		if currentSize == prevSize {
			stableDuration += checkInterval
		} else {
			stableDuration = 0 // 如果文件大小变化，重置稳定计时
		}

		prevSize = currentSize

		// 如果文件已经稳定超过 debounceDuration，则认为文件写入完成
		if stableDuration >= debounceDuration {
			fmt.Println("File write complete, processing:", filePath)
			handleFile(filePath)
			return
		}

		// 等待下一个时间间隔
		time.Sleep(checkInterval)
	}
}

// 文件处理函数
func handleFile(filePath string) {
	fmt.Println("Processing file:", filePath)
	// 执行你的操作
}
