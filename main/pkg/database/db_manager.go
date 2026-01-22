package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	bizctx "ttpos-server-go/pkg/context"
)

type DBManager struct {
	lock          *sync.Mutex
	dbs           map[uint64]*gorm.DB
	conf          config.DatabaseConf
	lastCheck     map[uint64]time.Time // 新增：记录最后检查时间
	checkInterval time.Duration        // 新增：检查间隔
}

var (
	instance *DBManager
	once     sync.Once
)

// GetDBManager 获取数据库管理器，如果是离线版，则0和商家库是同一个数据库
func GetDBManager(conf config.DatabaseConf) *DBManager {
	once.Do(func() {
		instance = &DBManager{
			dbs:           make(map[uint64]*gorm.DB),
			conf:          conf,
			lock:          &sync.Mutex{},
			lastCheck:     make(map[uint64]time.Time),
			checkInterval: 10 * time.Second,
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
		// 只有超过检查间隔才进行健康检查
		if time.Since(m.lastCheck[index]) > m.checkInterval {
			if sqlDB, err := db.DB(); err == nil {
				if err := sqlDB.Ping(); err != nil {
					// 先关闭现有连接
					if closeErr := sqlDB.Close(); closeErr != nil {
						log.Printf("关闭失效连接失败: %v", closeErr)
					}
					// 连接失效，删除并重建
					delete(m.dbs, index)
					delete(m.lastCheck, index)
				} else {
					// 更新检查时间
					m.lastCheck[index] = time.Now()
					return db
				}
			}
		} else {
			// 未到检查时间，直接返回
			return db
		}
	}

	dbName := "saas"
	if index > 0 {
		dbName = fmt.Sprintf("%s%d", constant.DBNamePrefix, index)
	}
	// 不存在，尝试连接
	companyDB, err := m.getConnection(m.conf, dbName) // 比如：shop1724054084 数据库 或 saas 库
	if err != nil {
		log.Printf("Error connecting to database for %s: %s\n", dbName, err)
		return nil
	}

	m.dbs[index] = companyDB
	m.lastCheck[index] = time.Now() // 记录检查时间
	return companyDB
}

// GetDBWithContext 返回绑定业务上下文的 *gorm.DB，确保数据库操作继承 trace/span 信息
func (m *DBManager) GetDBWithContext(ctx bizctx.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}
	db := m.GetDB(ctx.GetDbId())
	if db == nil {
		return nil
	}
	// 如果gin上下文存在，则使用gin上下文中的context
	if ginCtx := ctx.GetGin(); ginCtx != nil && ginCtx.Request != nil {
		if stdCtx := ginCtx.Request.Context(); stdCtx != nil {
			return db.WithContext(stdCtx)
		}
	}
	// 如果上下文存在，则使用上下文中的context
	if stdCtx := ctx.GetContext(); stdCtx != nil {
		return db.WithContext(stdCtx)
	}
	return db
}

func (m *DBManager) MustGetDB(idx uint64) (*gorm.DB, error) {
	db := m.GetDB(idx)
	if db == nil {
		return nil, fmt.Errorf("db err: %d not found", idx)
	}
	return db, nil
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

// CloseAll 关闭所有数据库连接
func (m *DBManager) CloseAll() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var errors []error
	for index, db := range m.dbs {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("关闭数据库连接 %d 失败: %w", index, err))
			}
		}
	}

	// 清空连接映射
	m.dbs = make(map[uint64]*gorm.DB)
	m.lastCheck = make(map[uint64]time.Time)

	if len(errors) > 0 {
		return fmt.Errorf("关闭数据库连接时发生错误: %v", errors)
	}

	return nil
}

// CloseAll 全局关闭所有数据库连接
func CloseAll() error {
	if instance != nil {
		return instance.CloseAll()
	}
	return nil
}
