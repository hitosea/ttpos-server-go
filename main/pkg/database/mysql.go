package database

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"ttpos-server-go/config"
	loggers "ttpos-server-go/pkg/logger"
)

func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	ignoreRecordNotFoundError := true // 忽略ErrRecordNotFound（记录未找到）错误
	logLevel := logger.Warn
	if config.Server.Mode == gin.DebugMode { // 调试模式
		ignoreRecordNotFoundError = false // 不忽略
		logLevel = logger.Info            // 详细日志
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		dbName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.TablePrefix, // 表名前缀
			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用自动创建外键约束
		SkipDefaultTransaction:                   true, // 禁用默认事务
		Logger: logger.New(
			&loggers.GormLog{Logger: loggers.SqlLogger}, // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold:             time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值
				LogLevel:                  logLevel,                                        // 日志级别
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  false,                                           // 彩色打印
			},
		),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	// 设置最大打开的连接数
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	// 设置连接的最大可存活时间
	sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Minute)

	return db, nil
}
