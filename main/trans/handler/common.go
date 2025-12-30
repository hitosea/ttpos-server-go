package handler

import (
	"fmt"
	"sync"
	"time"
	"ttpos-server-go/config"
	loggers "ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

//var sourceConf = config.DatabaseConf{
//	Host:          "localhost",
//	Port:          25443,
//	User:          "root",
//	Password:      "",
//	TablePrefix:   "jjjfood_",
//	SlowQueryTime: 0,
//}
//var sourceDBName = "shop1724054084"

var sourceConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          25443,
	User:          "root",
	Password:      "",
	TablePrefix:   "jjjfood_",
	SlowQueryTime: 0,
}
var sourceDBName = "shop1724054088"

var targetConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          13306,
	User:          "root",
	Password:      "",
	TablePrefix:   "ttpos_",
	SlowQueryTime: 0,
}
var targetDBName = "shop4477708931072000"

var SqlTransLogger *zap.Logger

func init() {
	SqlTransLogger = loggers.NewLoggerInstance(".trans")
}
func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	ignoreRecordNotFoundError := false // 忽略ErrRecordNotFound（记录未找到）错误
	logLevel := logger.Info
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		dbName,
	)
	// 初始化会话
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.TablePrefix, // 表名前缀
			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用自动创建外键约束
		SkipDefaultTransaction:                   true, // 禁用默认事务
		Logger: logger.New(
			&loggers.GormLog{Logger: SqlTransLogger}, // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold:             time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值
				LogLevel:                  logLevel,                                        // 日志级别
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  false,                                           // 彩色打印
			},
		),
	})
}

var SonyFlakeIdOnce sync.Once

func InitializeSonyFlakeId() {
	SonyFlakeIdOnce.Do(func() {
		utils.InitSonyFlakeId()
	})
}
