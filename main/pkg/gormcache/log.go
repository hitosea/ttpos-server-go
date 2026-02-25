package gormcache

import (
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// logDebug 安全地记录 debug 日志
func logDebug(msg string, fields ...zap.Field) {
	if logger.Logger != nil {
		logger.Logger.Debug(msg, fields...)
	}
}

// logInfo 安全地记录 info 日志
func logInfo(msg string, fields ...zap.Field) {
	if logger.Logger != nil {
		logger.Logger.Info(msg, fields...)
	}
}

// logWarn 安全地记录 warn 日志
func logWarn(msg string, fields ...zap.Field) {
	if logger.Logger != nil {
		logger.Logger.Warn(msg, fields...)
	}
}

// logError 安全地记录 error 日志
func logError(msg string, fields ...zap.Field) {
	if logger.Logger != nil {
		logger.Logger.Error(msg, fields...)
	}
}
