package main

//import (
//	"flag"
//	"fmt"
//	"log"
//	"os"
//	"path/filepath"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/fsnotify/fsnotify"
//)
//
//var (
//	fileEventTimes    sync.Map
//	debounceDuration1 = 1 * time.Second
//	debounceDuration  = 5 * time.Second        // 静默期，确保文件没有新的写入操作
//	checkInterval     = 500 * time.Millisecond // 文件大小检测间隔
//)
//
//func main() {
//	// 使用 flag 包解析命令行参数
//	watchDir := flag.String("watchDir", "", "Directory to watch for file changes")
//	flag.Parse()
//
//	// 确保提供了 watchDir 参数
//	if *watchDir == "" {
//		log.Fatal("Please specify a directory to watch using -watchDir")
//	}
//
//	watcher, err := fsnotify.NewWatcher()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer watcher.Close()
//
//	var wg sync.WaitGroup
//
//	go func() {
//		for {
//			select {
//			case event, ok := <-watcher.Events:
//				if !ok {
//					return
//				}
//
//				// 只处理非排除文件和文件夹的创建或写入事件
//				if (event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write) && !shouldIgnore(event.Name) {
//					fileInfo, err := os.Stat(event.Name)
//					if err == nil && fileInfo.IsDir() {
//						log.Println("New directory detected:", event.Name)
//						addWatcherRecursive(watcher, event.Name)
//					} else if shouldProcess(event.Name) {
//						wg.Add(1)
//						go checkIfFileComplete(event.Name, &wg) // 检测文件是否写入完成
//					}
//				}
//			case err, ok := <-watcher.Errors:
//				if !ok {
//					return
//				}
//				log.Println("Error:", err)
//			}
//		}
//	}()
//
//	addWatcherRecursive(watcher, *watchDir)
//
//	<-make(chan struct{})
//}
//
//// 递归添加目录及其子目录的监听，排除特定文件夹
//func addWatcherRecursive(watcher *fsnotify.Watcher, dir string) {
//	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//		// 排除以 .oss 开头的文件夹
//		if info.IsDir() && strings.HasPrefix(info.Name(), ".oss") {
//			log.Println("Skipping directory:", path)
//			return filepath.SkipDir
//		}
//		if info.IsDir() {
//			log.Println("Watching directory:", path)
//			return watcher.Add(path)
//		}
//		return nil
//	})
//	if err != nil {
//		log.Fatal("Error watching directories:", err)
//	}
//}
//
//// 判断是否应该忽略该文件或文件夹
//func shouldIgnore(path string) bool {
//	// 排除以 .temp 结尾的文件
//	if strings.HasSuffix(path, ".temp") {
//		log.Println("Ignoring file:", path)
//		return true
//	}
//	return false
//}
//
//// 判断是否应该处理该文件（基于时间窗口去重）
//func shouldProcess(filePath string) bool {
//	now := time.Now()
//	if lastProcessed, ok := fileEventTimes.Load(filePath); ok {
//		if time.Since(lastProcessed.(time.Time)) < debounceDuration1 {
//			return false
//		}
//	}
//	fileEventTimes.Store(filePath, now)
//	return true
//}
//
//// 检测文件是否已经写入完成
//func checkIfFileComplete(filePath string, wg *sync.WaitGroup) {
//	defer wg.Done()
//
//	var prevSize int64 = -1
//	var stableDuration time.Duration
//
//	for {
//		fileInfo, err := os.Stat(filePath)
//		if err != nil {
//			log.Println("Error reading file info:", err)
//			return
//		}
//
//		currentSize := fileInfo.Size()
//
//		// 如果文件大小在 checkInterval 时间内没有变化，则认为文件稳定
//		if currentSize == prevSize {
//			stableDuration += checkInterval
//		} else {
//			stableDuration = 0 // 如果文件大小变化，重置稳定计时
//		}
//
//		prevSize = currentSize
//
//		// 如果文件已经稳定超过 debounceDuration，则认为文件写入完成
//		if stableDuration >= debounceDuration {
//			fmt.Println("File write complete, processing:", filePath)
//			handleFile(filePath)
//			fileEventTimes.Delete(filePath) // 移除缓存记录
//			return
//		}
//
//		// 等待下一个时间间隔
//		time.Sleep(checkInterval)
//	}
//}
//
//// 文件处理函数
//func handleFile(filePath string) {
//	fmt.Println("Processing file:", filePath)
//	// 执行你的操作
//}
