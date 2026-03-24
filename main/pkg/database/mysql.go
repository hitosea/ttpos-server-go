package database

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"ttpos-server-go/config"
	loggers "ttpos-server-go/pkg/logger"
)

func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	ignoreRecordNotFoundError := true        // 忽略ErrRecordNotFound（记录未找到）错误
	if config.Server.Mode == gin.DebugMode { // 调试模式
		ignoreRecordNotFoundError = false // 不忽略
	}
	// 开启 interpolateParams 让驱动在客户端完成参数插值，避免 server-side prepared statements。
	// 在 MySQL Router 读写分离场景下，prepared statements 会让连接更容易被 pin 到 primary，
	// 从而导致读请求难以路由到 secondary；这里不要随意移除。
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		dbName,
	)

	// 使用自定义的GORM日志记录器
	customLogger := loggers.NewCustomGormLogger(
		loggers.SqlLogger,
		time.Duration(conf.SlowQueryTime)*time.Second, // 慢查询阈值
		ignoreRecordNotFoundError,                     // 忽略ErrRecordNotFound错误
		dbName,                                        // 数据库名
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.TablePrefix, // 表名前缀
			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
		},
		DisableForeignKeyConstraintWhenMigrating: true,         // 禁用自动创建外键约束
		SkipDefaultTransaction:                   true,         // 禁用默认事务
		Logger:                                   customLogger, // 使用自定义日志记录器
	})
	if err != nil {
		return nil, err
	}
	enableGormTracing(db, dbName)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	// 设置最大打开的连接数
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	// 设置连接的最大可存活时间
	sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)

	return db, nil
}
