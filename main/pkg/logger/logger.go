package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"ttpos-server-go/config"
)

var Logger *zap.Logger

// Init 初始化日志
func Init() error {
	// 创建日志目录
	if err := os.MkdirAll(config.Log.Dir, 0755); err != nil {
		return fmt.Errorf("can't create log directory: %v", err)
	}

	// 获取当前日期作为日志文件名
	now := time.Now()
	logfile := filepath.Join(config.Log.Dir, fmt.Sprintf("%s.log", now.Format(time.DateOnly)))

	// 配置 lumberjack
	hook := &lumberjack.Logger{
		Filename:   logfile,              // 日志文件路径
		MaxSize:    config.Log.MaxSize,   // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: config.Log.MaxBackup, // 保留最近14天的日志文件
		MaxAge:     config.Log.MaxBackup, // 保留最近14天的日志文件
		Compress:   true,                 // 是否压缩
		LocalTime:  true,                 // 使用本地时间
	}

	// 编码器配置
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

	levelMap := map[string]zapcore.Level{
		"debug": zap.DebugLevel,
		"info":  zap.InfoLevel,
		"warn":  zap.WarnLevel,
		"error": zap.ErrorLevel,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(levelMap[config.Log.Level])

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		zapcore.AddSync(hook),                 // 输出到文件
		atomicLevel,                           // 日志级别
	)

	// 创建Logger
	Logger = zap.New(core,
		zap.AddCaller(),                   // 调用文件和行号
		zap.AddStacktrace(zap.ErrorLevel), // Error级别日志才会显示堆栈信息
	)

	// 定时每天清理一次日志
	startCron()
	return nil
}

func startCron() {
	c := cron.New()
	_, _ = c.AddFunc(config.Log.CleanSchedule, func() {
		if err := cleanOldLogs(); err != nil {
			Logger.Error("failed to clean old logs", zap.Error(err))
		}
	})
	c.Start()
}

// cleanOldLogs 清理旧日志文件
func cleanOldLogs() error {
	threshold := time.Now().AddDate(0, 0, -14) // 14天前的时间

	return filepath.Walk(config.Log.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(threshold) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old log file %s: %v", path, err)
			}
		}

		return nil
	})
}

type GormLog struct {
	Logger *zap.Logger
}

func (log *GormLog) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	msg = strings.ReplaceAll(msg, "\n", " ")
	log.Logger.Info(msg)
}
