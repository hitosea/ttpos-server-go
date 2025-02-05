package repository

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	ignoreRecordNotFoundError := true // 忽略ErrRecordNotFound（记录未找到）错误
	logLevel := logger.Warn
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
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold:             time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值
				LogLevel:                  logLevel,                                        // 日志级别
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  true,                                            // 彩色打印
			},
		),
	})
}

func TestCreateMultiLanguageName(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "QWERASDFQWE23421",
		RootPassword:  "QWERASDFQWE23421",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop1111000")
	if err != nil {
		panic(err)
	}
	multiLanguageNameRepository := NewMultiLanguageNameRepository(db)

	multiLanguageNames := []model.MultiLanguageName{
		{
			EnName:   "Signature Menu",
			ZhName:   "招牌菜单",
			ZhTwName: "招牌菜單",
		},
		{
			EnName:   "Dessert Menu",
			ZhName:   "甜点菜单",
			ZhTwName: "甜點菜單",
		},
		{
			EnName:   "Beverage Menu",
			ZhName:   "饮料菜单",
			ZhTwName: "飲料菜單",
		},
		{
			EnName:   "Appetizer Menu",
			ZhName:   "开胃菜菜单",
			ZhTwName: "開胃菜菜單",
		},
	}

	for _, multiLanguageName := range multiLanguageNames {
		_, err := multiLanguageNameRepository.CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			panic(err)
		}
	}
}
