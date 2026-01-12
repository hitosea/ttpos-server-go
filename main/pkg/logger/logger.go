package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"

	"ttpos-server-go/config"
)

var Logger *zap.Logger
var SqlLogger *zap.Logger

// Init 初始化日志
func Init() error {
	// 创建日志目录
	if err := os.MkdirAll(config.Log.Dir, 0755); err != nil {
		return fmt.Errorf("can't create log directory: %v", err)
	}

	// 创建Logger
	Logger = NewLoggerInstance("")
	SqlLogger = NewLoggerInstance(".sql")
	// 定时每天清理一次日志
	startCron()
	return nil
}

func NewLoggerInstance(name string) *zap.Logger {
	now := time.Now()
	logFile := filepath.Join(config.Log.Dir, fmt.Sprintf("%s%s.log", now.Format(time.DateOnly), name))

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

	// 配置 lumberjack
	hook := &lumberjack.Logger{
		Filename:   logFile,              // 日志文件路径
		MaxSize:    config.Log.MaxSize,   // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: config.Log.MaxBackup, // 保留最近14天的日志文件
		MaxAge:     config.Log.MaxBackup, // 保留最近14天的日志文件
		Compress:   true,                 // 是否压缩
		LocalTime:  true,                 // 使用本地时间
	}

	levelMap := map[string]zapcore.Level{
		"debug": zap.DebugLevel,
		"info":  zap.InfoLevel,
		"warn":  zap.WarnLevel,
		"error": zap.ErrorLevel,
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(levelMap[config.Log.Level])

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		zapcore.AddSync(hook),                 // 输出到文件
		atomicLevel,                           // 日志级别
	)

	return zap.New(core,
		zap.AddCaller(),                   // 调用文件和行号
		zap.AddStacktrace(zap.ErrorLevel), // Error级别日志才会显示堆栈信息
	)
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

func (log *GormLog) Printf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	msg = strings.ReplaceAll(msg, "\n", " ")
	log.Logger.Info(msg)
}

// CustomGormLogger 自定义GORM日志记录器
// 只在调用Debug()时或者报错时打印SQL日志
type CustomGormLogger struct {
	ZapLogger                 *zap.Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  logger.LogLevel // 添加日志级别字段
	DBName                    string          // 数据库名
}

// NewCustomGormLogger 创建自定义GORM日志记录器
func NewCustomGormLogger(zapLogger *zap.Logger, slowThreshold time.Duration, ignoreRecordNotFoundError bool, dbName string) *CustomGormLogger {
	return &CustomGormLogger{
		ZapLogger:                 zapLogger,
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: ignoreRecordNotFoundError,
		LogLevel:                  logger.Silent, // 默认为Silent模式
		DBName:                    dbName,
	}
}

// LogMode 设置日志模式
func (l *CustomGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level // 设置新的日志级别
	return &newLogger
}

// Info 信息日志（我们不记录，除非是Debug模式）
func (l *CustomGormLogger) Info(ctx context.Context, msg string, data ...any) {
	// 只有在Debug模式下才记录Info日志
	if l.isDebugMode() {
		l.ZapLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 警告日志
func (l *CustomGormLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.ZapLogger.Warn(fmt.Sprintf(msg, data...))
}

// Error 错误日志
func (l *CustomGormLogger) Error(ctx context.Context, msg string, data ...any) {
	l.ZapLogger.Error(fmt.Sprintf(msg, data...))
}

// Trace 跟踪SQL执行
func (l *CustomGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 格式化SQL语句
	formattedSQL := strings.ReplaceAll(sql, "\n", " ")
	formattedSQL = strings.ReplaceAll(formattedSQL, "\t", " ")

	// 构建日志字段
	fields := []zap.Field{
		zap.String("database", l.DBName),
		zap.String("sql", formattedSQL),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil:
		// 有错误时总是记录（除非是忽略的RecordNotFound错误）
		if l.IgnoreRecordNotFoundError && err.Error() == "record not found" {
			return
		}
		l.ZapLogger.Error("SQL执行出错", append(fields, zap.Error(err))...)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		// 慢查询总是记录
		l.ZapLogger.Warn("慢查询检测", append(fields, zap.String("threshold", l.SlowThreshold.String()))...)

	case l.isDebugMode():
		// Debug模式下记录所有SQL
		l.ZapLogger.Info("SQL执行", fields...)

		// 默认情况下不记录正常的SQL执行
	}
}

// isDebugMode 检查是否为Debug模式
// 只有在显式调用Debug()时才返回true
func (l *CustomGormLogger) isDebugMode() bool {
	return l.LogLevel == logger.Info || config.Server.Mode == gin.DebugMode
}
