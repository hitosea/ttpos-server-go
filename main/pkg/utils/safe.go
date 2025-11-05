package utils

import (
	"runtime/debug"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// SafeGo 安全地执行goroutine
func SafeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 打印panic信息和堆栈
				logger.Logger.Error("SafeGo: panic ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
			}
		}()
		f()
	}()
}
