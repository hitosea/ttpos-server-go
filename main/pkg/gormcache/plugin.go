package gormcache

import (
	"encoding/json"
	"reflect"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// queryType 查询类型
type queryType string

const (
	queryTypeQuery  queryType = "query"
	queryTypeCreate queryType = "create"
	queryTypeUpdate queryType = "update"
	queryTypeDelete queryType = "delete"
)

// Plugin GORM 缓存插件
// 实现了 gorm.Plugin 接口
type Plugin struct {
	conf *Config

	// queue 用于 Easer 请求合并的队列
	queue sync.Map

	// callbacks 保存原始 GORM 回调
	callbacks map[queryType]func(db *gorm.DB)

	// tableWhitelist 表白名单（空则缓存所有）
	tableWhitelist map[string]struct{}

	// tableBlacklist 表黑名单
	tableBlacklist map[string]struct{}

	// dbBlacklist 数据库黑名单（商户级别禁用缓存）
	dbBlacklist map[string]struct{}
}

// New 创建缓存插件实例
func New(conf *Config) *Plugin {
	if conf == nil {
		conf = DefaultConfig()
	}

	p := &Plugin{
		conf:           conf,
		callbacks:      make(map[queryType]func(db *gorm.DB)),
		tableWhitelist: make(map[string]struct{}),
		tableBlacklist: make(map[string]struct{}),
		dbBlacklist:    make(map[string]struct{}),
	}

	// 构建表白名单
	for _, t := range conf.Tables {
		p.tableWhitelist[t] = struct{}{}
	}

	// 构建表黑名单
	for _, t := range conf.ExcludeTables {
		p.tableBlacklist[t] = struct{}{}
	}

	// 构建数据库黑名单（商户级别禁用）
	for _, db := range conf.ExcludeDBs {
		p.dbBlacklist[db] = struct{}{}
	}

	return p
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "gorm:table-cache"
}

// Initialize 初始化插件
func (p *Plugin) Initialize(db *gorm.DB) error {
	// 1. 保存原始 Query 回调
	p.callbacks[queryTypeQuery] = db.Callback().Query().Get("gorm:query")
	p.callbacks[queryTypeCreate] = db.Callback().Create().Get("gorm:create")
	p.callbacks[queryTypeUpdate] = db.Callback().Update().Get("gorm:update")
	p.callbacks[queryTypeDelete] = db.Callback().Delete().Get("gorm:delete")

	// 2. 替换 Query 回调
	if err := db.Callback().Query().Replace("gorm:query", p.queryCallback); err != nil {
		return err
	}

	// 3. 注册写操作的缓存失效回调
	if p.conf.Cacher != nil && (p.conf.InvalidateOnWrite == nil || *p.conf.InvalidateOnWrite) {
		// Create 后失效
		if err := db.Callback().Create().After("gorm:create").Register(
			"gorm:cache:invalidate:create",
			p.invalidateCallback(queryTypeCreate),
		); err != nil {
			return err
		}

		// Update 后失效
		if err := db.Callback().Update().After("gorm:update").Register(
			"gorm:cache:invalidate:update",
			p.invalidateCallback(queryTypeUpdate),
		); err != nil {
			return err
		}

		// Delete 后失效
		if err := db.Callback().Delete().After("gorm:delete").Register(
			"gorm:cache:invalidate:delete",
			p.invalidateCallback(queryTypeDelete),
		); err != nil {
			return err
		}
	}

	logInfo("gormcache: 插件初始化完成",
		zap.Bool("easer", p.conf.Easer),
		zap.Bool("cacher", p.conf.Cacher != nil),
		zap.Int("tableWhitelist", len(p.tableWhitelist)),
		zap.Int("tableBlacklist", len(p.tableBlacklist)),
		zap.Int("dbBlacklist", len(p.dbBlacklist)),
	)

	return nil
}

// queryCallback 查询回调（核心逻辑）
// 参考 go-gorm/caches 官方实现
func (p *Plugin) queryCallback(db *gorm.DB) {
	// 检查是否跳过缓存
	if p.shouldSkipCache(db) {
		p.executeQuery(db)
		return
	}

	// 构建缓存标识（使用 callbacks.BuildQuerySQL 构建 SQL）
	// 注意：BuildQuerySQL 会检查 SQL 是否已构建，不会重复构建
	tableName, key := buildIdentifier(db)

	// 检查表是否需要缓存
	if !p.shouldCacheTable(tableName) {
		p.executeQuery(db)
		return
	}

	// 1. 尝试从缓存获取
	if p.conf.Cacher != nil {
		if result, found := p.conf.Cacher.Get(db.Statement.Context, key); found {
			// 缓存命中，回填结果
			if err := p.replaceResult(db, result); err == nil {
				if p.conf.Debug {
					logDebug("gormcache: 缓存命中",
						zap.String("table", tableName),
						zap.String("key", key),
					)
				}
				return
			}
		}
	}

	// 2. 缓存未命中，执行查询
	if p.conf.Easer {
		// 使用 Easer 合并相同请求
		task := &queryTask{
			id: key,
			queryCb: func() {
				p.executeQuery(db)
			},
		}
		ease(task, &p.queue)
	} else {
		p.executeQuery(db)
	}

	// 3. 将结果存入缓存
	if p.conf.Cacher != nil && db.Error == nil {
		p.storeInCache(db, tableName, key)
	}
}

// executeQuery 执行原始查询
func (p *Plugin) executeQuery(db *gorm.DB) {
	if cb, ok := p.callbacks[queryTypeQuery]; ok && cb != nil {
		cb(db)
	}
}

// invalidateCallback 返回写操作的失效回调
func (p *Plugin) invalidateCallback(qt queryType) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Error != nil {
			return
		}

		tableName := extractTableNameFromMutator(db)
		if tableName == "" || tableName == "unknown" {
			return
		}

		// 检查表是否在缓存白名单中，不在则跳过（避免无效的 Redis 请求）
		if !p.shouldCacheTable(tableName) {
			return
		}

		// 构建表索引键（多租户隔离）
		tableKey := buildTableKey(db, tableName)

		// 失效该表的所有缓存
		if err := p.conf.Cacher.InvalidateTable(db.Statement.Context, tableKey); err != nil {
			logWarn("gormcache: 缓存失效失败",
				zap.String("tableKey", tableKey),
				zap.String("operation", string(qt)),
				zap.Error(err),
			)
		} else if p.conf.Debug {
			logDebug("gormcache: 缓存已失效",
				zap.String("tableKey", tableKey),
				zap.String("operation", string(qt)),
			)
		}
	}
}

// shouldSkipCache 检查是否应该跳过缓存
func (p *Plugin) shouldSkipCache(db *gorm.DB) bool {
	// 检查上下文中的跳过标记
	if skip, ok := db.Statement.Context.Value(skipCacheKey{}).(bool); ok && skip {
		return true
	}

	// 没有配置缓存器且没有启用 Easer
	if p.conf.Cacher == nil && !p.conf.Easer {
		return true
	}

	// 检查数据库是否在黑名单中（商户级别禁用）
	if len(p.dbBlacklist) > 0 {
		dbName := extractDBName(db)
		if _, ok := p.dbBlacklist[dbName]; ok {
			return true
		}
	}

	return false
}

// shouldCacheTable 检查表是否需要缓存
func (p *Plugin) shouldCacheTable(tableName string) bool {
	// 如果设置了白名单，只缓存白名单中的表
	if len(p.tableWhitelist) > 0 {
		_, ok := p.tableWhitelist[tableName]
		return ok
	}

	// 检查黑名单
	if _, ok := p.tableBlacklist[tableName]; ok {
		return false
	}

	return true
}

// storeInCache 将查询结果存入缓存
func (p *Plugin) storeInCache(db *gorm.DB, tableName string, key string) {
	// 检查结果行数
	if p.conf.MaxRows > 0 && db.Statement.RowsAffected > p.conf.MaxRows {
		if p.conf.Debug {
			logDebug("gormcache: 结果行数超过限制，跳过缓存",
				zap.String("table", tableName),
				zap.Int64("rows", db.Statement.RowsAffected),
				zap.Int64("maxRows", p.conf.MaxRows),
			)
		}
		return
	}

	// 构建缓存结果
	result := &QueryResult{
		Dest:         db.Statement.Dest,
		RowsAffected: db.Statement.RowsAffected,
	}

	// 存入缓存
	ttl := p.conf.TTL
	if ttl <= 0 {
		ttl = DefaultConfig().TTL
	}

	// 构建表索引键（多租户隔离）
	tableKey := buildTableKey(db, tableName)

	if err := p.conf.Cacher.Store(db.Statement.Context, tableKey, key, result, ttl); err != nil {
		logWarn("gormcache: 缓存存储失败",
			zap.String("tableKey", tableKey),
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

// replaceResult 将缓存结果回填到 GORM Statement
func (p *Plugin) replaceResult(db *gorm.DB, result *QueryResult) error {
	if result == nil {
		return nil
	}

	// 回填 RowsAffected
	db.Statement.RowsAffected = result.RowsAffected

	// 回填 Dest
	if result.Dest != nil && db.Statement.Dest != nil {
		return copyValue(db.Statement.Dest, result.Dest)
	}

	return nil
}

// copyValue 将源值复制到目标（通过 JSON 序列化/反序列化）
func copyValue(dest any, src any) error {
	// 序列化源值
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	// 反序列化到目标
	return json.Unmarshal(data, dest)
}

// setPointedValue 使用反射设置指针值
func setPointedValue(dest any, src any) {
	destVal := reflect.ValueOf(dest)
	srcVal := reflect.ValueOf(src)

	if destVal.Kind() == reflect.Ptr && srcVal.Kind() == reflect.Ptr {
		destVal.Elem().Set(srcVal.Elem())
	}
}

// ============ 上下文控制 ============

// skipCacheKey 跳过缓存的上下文键
type skipCacheKey struct{}

// SkipCache 返回跳过缓存的 GORM Scope
// 用法: db.Scopes(gormcache.SkipCache()).Find(&users)
func SkipCache() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.WithContext(
			db.Statement.Context,
		).Set("gorm:cache:skip", true)
	}
}
