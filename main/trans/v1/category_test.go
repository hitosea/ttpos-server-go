//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
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

func TestCategory(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	categoryService := CategoryService{db: db}
	categoryList, err := categoryService.GetCategoryList()
	if err != nil {
		panic(err)
	}
	jsonData, err := json.Marshal(categoryList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}

func TestConvertCategory(t *testing.T) {
	testConvertCategory()
}

func testConvertCategory() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDb, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	categoryService := CategoryService{db: db, targetDB: targetDb}
	err = categoryService.ConvertCategory()
	if err != nil {
		panic(err)
	}
}
