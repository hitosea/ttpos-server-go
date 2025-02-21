package database

import (
	"fmt"
	"log"
	"sync"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
)

type DBManager struct {
	lock *sync.Mutex
	dbs  map[uint64]*gorm.DB
	conf config.DatabaseConf
}

var (
	instance *DBManager
	once     sync.Once
)

// GetDBManager 获取数据库管理器，如果是离线版，则0和商家库是同一个数据库
func GetDBManager(conf config.DatabaseConf) *DBManager {
	once.Do(func() {
		instance = &DBManager{
			dbs:  make(map[uint64]*gorm.DB),
			conf: conf,
			lock: &sync.Mutex{},
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
	if err := db.Scopes(repository.NotDeleted).Debug().Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies: %s", err)
	}
	for _, company := range companies {
		companyDB, err := m.getConnection(conf, fmt.Sprintf("%s%d", constant.DBNamePrefix, company.Uuid)) // 比如：shop1724054084 数据库
		if err != nil {
			log.Fatalf("Error connecting to database for company %d: %s", company.Uuid, err)
		}
		m.dbs[company.Uuid] = companyDB
	}
}

// getConnection 获取数据库连接
func (m *DBManager) getConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	switch conf.DBType {
	case "mysql":
		return NewMySQLConnection(conf, dbName)
	case "sqlite":
		return NewSQLiteConnection(conf)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", conf.DBType)
	}
}

// GetDB 获取数据库，动态获取新增商家
func (m *DBManager) GetDB(index uint64) *gorm.DB {
	m.lock.Lock()
	defer m.lock.Unlock()
	if db, ok := m.dbs[index]; ok {
		return db
	}
	// 不存在，尝试连接
	companyDB, err := m.getConnection(m.conf, fmt.Sprintf("%s%d", constant.DBNamePrefix, index)) // 比如：shop1724054084 数据库
	if err != nil {
		log.Printf("Error connecting to database for company %d: %s\n", index, err)
		return nil
	}
	m.dbs[index] = companyDB
	return companyDB
}

func (m *DBManager) GetDBNameList() map[uint64]string {
	dbNames := make(map[uint64]string)
	for dbName := range m.dbs {
		dbNames[dbName] = fmt.Sprintf("%s%d", constant.DBNamePrefix, dbName)
	}
	return dbNames
}

// SetMockDB 添加测试用的DB实例
func (m *DBManager) SetMockDB(db *gorm.DB) {
	if m.dbs == nil {
		m.dbs = make(map[uint64]*gorm.DB)
	}
	m.dbs[constant.MockDB] = db
}

// GetMockDB 获取测试用的DB实例
func (m *DBManager) GetMockDB() *gorm.DB {
	return m.dbs[constant.MockDB]
}
