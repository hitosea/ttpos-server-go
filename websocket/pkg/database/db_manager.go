package database

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"

	"websocket/config"
	"websocket/constant"
	"websocket/model"
)

type DBManager struct {
	lock *sync.Mutex
	dbs  map[uint64]*gorm.DB
	conf config.DatabaseConf
}

var (
	Instance *DBManager
	once     sync.Once
)

// GetDBManager 获取数据库管理器，如果是离线版，则0和商家库是同一个数据库
func GetDBManager(conf config.DatabaseConf) *DBManager {
	once.Do(func() {
		Instance = &DBManager{
			dbs:  make(map[uint64]*gorm.DB),
			conf: conf,
			lock: &sync.Mutex{},
		}
		Instance.initDBs(conf)
	})
	return Instance
}

// initDBs 初始化数据库
func (m *DBManager) initDBs(conf config.DatabaseConf) {
	if m.dbs == nil {
		m.dbs = make(map[uint64]*gorm.DB)
	}
	// 主数据库
	db, err := m.getConnection(conf, conf.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	m.dbs[constant.DefaultDB] = db
	// 根据 APP 表实例化数据库连接
	var companies []model.Company
	if err := db.Where("delete_time = ?", 0).Debug().Find(&companies).Error; err != nil {
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
	return NewMySQLConnection(conf, dbName)
}

// GetDB 获取数据库
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
	}
	m.dbs[index] = companyDB
	return companyDB
}

func (m *DBManager) GetDBNameList() map[uint64]string {
	dbNames := make(map[uint64]string)
	for dbName := range m.dbs {
		dbNames[dbName] = fmt.Sprintf("%s%d", "shop", dbName)
	}
	return dbNames
}
