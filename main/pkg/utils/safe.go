package utils

import (
	"log"
	"runtime/debug"
)

// SafeGo 安全地执行goroutine
func SafeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 打印panic信息和堆栈
				log.Printf("[PANIC] %v\n%s \n", r, debug.Stack())
			}
		}()
		f()
	}()
}
