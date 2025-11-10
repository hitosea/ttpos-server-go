package utils

import (
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ExecuteSafeGoroutine 以安全模式启动一个协程，内部会捕获 panic 并记录日志。
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 捕获 panic 并记录日志
				logger.Logger.Error("独立协程发生 panic", zap.Any("panic_value", r), zap.Stack("stack_trace"))
			}
		}()
		fn()
	}()
}
