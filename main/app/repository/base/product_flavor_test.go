//go:build integration

package base

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
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

func TestCreateProductFlavor(t *testing.T) {
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
	productFlavorRepository := NewProductFlavorRepoImpl(db)

	productFlavors := []model.ProductFlavor{
		{
			Name: "大杯",
			MultiLanguageName: model.MultiLanguageName{
				EnName:   "Large Cup",
				ZhName:   "大杯",
				ZhTwName: "大杯",
			},
		},
		{
			Name: "小杯",
			MultiLanguageName: model.MultiLanguageName{
				EnName:   "Small Cup",
				ZhName:   "小杯",
				ZhTwName: "小杯",
			},
		},
	}

	for _, productFlavor := range productFlavors {
		_, err := productFlavorRepository.CreateProductFlavor(productFlavor)
		if err != nil {
			panic(err)
		}
	}
}

func TestGetProductFlavorList(t *testing.T) {
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
	productFlavorRepository := NewProductFlavorRepo(db)
	productFlavors, err := productFlavorRepository.GetProductFlavorList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productFlavors)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))

}

func NewSQLiteConnection(conf config.DatabaseConf) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(conf.Database), &gorm.Config{
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
}

func TestCreateProductFlavorListSQLite(t *testing.T) {
	db, err := NewSQLiteConnection(config.DatabaseConf{
		Database: "test.db",
	})
	if err != nil {
		panic(err)
	}
	productFlavorRepository := NewProductFlavorRepo(db)
	db.AutoMigrate(&model.ProductFlavor{}, &model.MultiLanguageName{})
	productFlavors := []model.ProductFlavor{
		{
			Name: "大杯",
			MultiLanguageName: model.MultiLanguageName{
				EnName:   "Large Cup",
				ZhName:   "大杯",
				ZhTwName: "大杯",
			},
		},
		{
			Name: "小杯",
			MultiLanguageName: model.MultiLanguageName{
				EnName:   "Small Cup",
				ZhName:   "小杯",
				ZhTwName: "小杯",
			},
		},
	}

	for _, productFlavor := range productFlavors {
		_, err := productFlavorRepository.CreateProductFlavor(productFlavor)
		if err != nil {
			panic(err)
		}
	}

	productFlavors, errGet := productFlavorRepository.GetProductFlavorList()
	if errGet != nil {
		panic(errGet)
	}
	json, err := json.Marshal(productFlavors)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}
