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

// initDBs 初始化数据库
func (m *DBManager) initDBs(conf config.DatabaseConf) {
	// 主数据库
	db, err := m.getConnection(conf, conf.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	m.dbs[constant.DefaultDB] = db
	// 根据 APP 表实例化数据库连接
	var companies []model.Company
	if err := db.Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies: %s", err)
	}
	for _, app := range companies {
		appDB, err := m.getConnection(conf, fmt.Sprintf("%s%d", constant.DBNamePrefix, app.ID)) // 比如：shop1724054084 数据库
		if err != nil {
			log.Fatalf("Error connecting to database for app %d: %s", app.ID, err)
		}
		m.dbs[app.ID] = appDB
	}
}

// getConnection 获取数据库连接
func (m *DBManager) getConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	switch conf.DBType {
	case "mysql":
		return NewMySQLConnection(conf, dbName)
	case "postgres":
		return NewPostgreSQLConnection(conf, dbName)
	case "sqlite":
		return NewSQLiteConnection(conf)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", conf.DBType)
	}
}

// GetDB 获取数据库
func (m *DBManager) GetDB(index uint) *gorm.DB {
	if db, ok := m.dbs[index]; ok {
		return db
	}
	panic(fmt.Sprintf("Database with index %d not found", index))
}

func (m *DBManager) GetDBNameList() map[uint]string {
	dbNames := make(map[uint]string)
	for dbName := range m.dbs {
		dbNames[dbName] = fmt.Sprintf("%s%d", constant.DBNamePrefix, dbName)
	}
	return dbNames
}

// SetMockDB 添加测试用的DB实例
func (m *DBManager) SetMockDB(db *gorm.DB) {
	if m.dbs == nil {
		m.dbs = make(map[uint]*gorm.DB)
	}
	m.dbs[constant.MockDB] = db
}

// GetMockDB 获取测试用的DB实例
func (m *DBManager) GetMockDB() *gorm.DB {
	return m.dbs[constant.MockDB]
}
