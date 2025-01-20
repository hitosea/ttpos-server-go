package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewSQLiteConnection(db string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(db), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "jjjfood_", // 表名前缀z
			SingularTable: true,       // 使用单一表名, eg. `User` => `user`
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold: time.Second, // 慢查询阈值设置为1秒
				LogLevel:      logger.Info, // 日志级别设置为Info，记录所有SQL
				Colorful:      true,        // 彩色打印
			},
		),
	})
}
