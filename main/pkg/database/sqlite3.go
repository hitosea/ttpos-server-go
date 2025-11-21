package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"ttpos-server-go/config"
)

func NewSQLiteConnection(conf config.DatabaseConf) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(conf.Database), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.TablePrefix, // 表名前缀z
			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold: time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值设置为1秒
				LogLevel:      logger.Info,                                     // 日志级别设置为Info，记录所有SQL
				Colorful:      true,                                            // 彩色打印
			},
		),
	})
	if err != nil {
		return nil, err
	}
	enableGormTracing(db, conf.Database)
	return db, nil
}
