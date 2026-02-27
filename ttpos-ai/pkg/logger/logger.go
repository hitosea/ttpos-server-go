package logger

import (
	"os"
	"path/filepath"
	"ttpos-ai/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.Logger

// Init 初始化日志
func Init() error {
	cfg := &config.Log

	// 确保日志目录存在
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return err
	}

	// 日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件输出
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "ai.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackup,
		MaxAge:     30,
		Compress:   true,
	}

	// 创建core
	var cores []zapcore.Core

	// 文件core
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)
	cores = append(cores, fileCore)

	// Debug模式下同时输出到控制台
	if config.Server.Mode == "debug" {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 组合core
	core := zapcore.NewTee(cores...)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))

	return nil
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
