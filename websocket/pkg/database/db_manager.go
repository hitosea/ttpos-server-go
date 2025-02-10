package database

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"

	"websocket/config"
	"websocket/model"
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

// initDBs 初始化数据库
func (m *DBManager) initDBs(conf config.DatabaseConf) {
	// Initialize the dbs map if it's nil
	if m.dbs == nil {
		m.dbs = make(map[uint]*gorm.DB)
	}
	// 主数据库
	if m.dbs[0] == nil {
		db, err := m.getConnection(conf, conf.Database)
		if err != nil {
			log.Fatalf("Error connecting to database: %s", err)
		}
		m.dbs[0] = db
	}
	// 根据 APP 表实例化数据库连接
	var companies []model.Company
	if err := m.dbs[0].Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies: %s", err)
	}
	for _, app := range companies {
		if _, ok := m.dbs[app.Uuid]; !ok {
			appDB, err := m.getConnection(conf, fmt.Sprintf("%s%d", "shop", app.Uuid)) // 比如：shop1724054084 数据库
			if err != nil {
				log.Fatalf("Error connecting to database for app %d: %s", app.Uuid, err)
			}
			m.dbs[app.Uuid] = appDB
		}
	}
}

// getConnection 获取数据库连接
func (m *DBManager) getConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	switch conf.DBType {
	case "mysql":
		return NewMySQLConnection(conf, dbName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", conf.DBType)
	}
}

// GetDB 获取数据库
func (m *DBManager) GetDB(index uint) *gorm.DB {
	if db, ok := m.dbs[index]; ok {
		return db
	}
	// 初始化数据库
	m.initDBs(config.Database)
	// 应用数据库
	if db, ok := m.dbs[index]; ok {
		return db
	}
	//
	panic(fmt.Sprintf("Database with index %d not found", index))
}

func (m *DBManager) GetDBNameList() map[uint]string {
	dbNames := make(map[uint]string)
	for dbName := range m.dbs {
		dbNames[dbName] = fmt.Sprintf("%s%d", "shop", dbName)
	}
	return dbNames
}
