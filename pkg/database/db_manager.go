package database

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
)

type DBManager struct {
	dbs map[uint]*gorm.DB
}

var (
	instance *DBManager
	once     sync.Once
)

// GetDBManager 获取数据库管理器，如果是离线版，则0和商家库是同一个数据库
func GetDBManager(conf config.DatabaseConf) *DBManager {
	once.Do(func() {
		instance = &DBManager{
			dbs: make(map[uint]*gorm.DB),
		}
		instance.initDBs(conf)
	})
	return instance
}

func (m *DBManager) initDBs(conf config.DatabaseConf) {
	// 主数据库
	db, err := NewMySQLConnection(conf, conf.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	m.dbs[constant.DefaultDB] = db
	// 根据 APP 表实例化数据库连接
	var apps []model.App
	if err := db.Find(&apps).Error; err != nil {
		log.Fatalf("Error querying apps: %s", err)
	}
	for _, app := range apps {
		appDB, err := NewMySQLConnection(conf, fmt.Sprintf("shop%d", app.AppId)) // 比如：shop1724054084 数据库
		if err != nil {
			log.Fatalf("Error connecting to database for app %d: %s", app.AppId, err)
		}
		m.dbs[app.AppId] = appDB
	}
}

func (m *DBManager) GetDB(index uint) *gorm.DB {
	if db, ok := m.dbs[index]; ok {
		return db
	}
	panic(fmt.Sprintf("Database with index %d not found", index))
}
