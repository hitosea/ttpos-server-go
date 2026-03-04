package gormcache

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"

	"ttpos-server-go/pkg/logger"
)

const (
	// IdentifierPrefix 缓存键前缀
	IdentifierPrefix = "gorm:cache:"
)

// buildIdentifier 构建缓存键（参考 go-gorm/caches 官方实现）
// 使用 callbacks.BuildQuerySQL 构建 SQL，然后生成缓存 key
// 格式: gorm:cache:{dbName}:{tableName}:{sql}:{params}
//
// 注意：callbacks.BuildQuerySQL 会检查 SQL 是否已构建，不会重复构建
// 因此可以安全地在 Query 回调中调用
//
// 重要：必须包含 dbName 以区分不同商户数据库，避免多租户数据串库
func buildIdentifier(db *gorm.DB, dbName string) (tableName string, key string) {
	// 使用 GORM 官方的 callbacks 包构建 SQL
	// 这是 go-gorm/caches 官方库使用的方法
	callbacks.BuildQuerySQL(db)

	tableName = extractTableName(db)

	// 构建唯一标识（包含数据库名以支持多租户隔离）
	var builder strings.Builder
	builder.WriteString(IdentifierPrefix)
	builder.WriteString(dbName)
	builder.WriteString(":")
	builder.WriteString(tableName)
	builder.WriteString(":")
	builder.WriteString(db.Statement.SQL.String())
	builder.WriteString(":")
	builder.WriteString(valuesToString(db.Statement.Vars))

	key = builder.String()
	return tableName, key
}

// buildTableKey 构建表索引键，格式为 {dbName}:{tableName}
// 用于多租户隔离，确保不同商户的缓存不会混淆
func buildTableKey(dbName string, tableName string) string {
	return dbName + ":" + tableName
}

// dbNameCache 缓存数据库名（用于非 CustomGormLogger 的备选方案）
// key: db.Config 指针地址, value: 数据库名
var dbNameCache sync.Map

// extractDBName 从 GORM DB 获取数据库名称
// 用于多租户隔离，确保不同商户的缓存不会混淆
// 优先从 CustomGormLogger.DBName 获取，备选通过 SQL 查询
func extractDBName(db *gorm.DB) string {
	if db.Config == nil {
		return "unknown"
	}

	// 方式1: 从 CustomGormLogger 获取数据库名（推荐，直接从内存读取）
	if db.Config.Logger != nil {
		if customLogger, ok := db.Config.Logger.(*logger.CustomGormLogger); ok {
			return customLogger.DBName
		}
	}

	// 方式2: 从缓存获取（避免重复查询）
	cacheKey := reflect.ValueOf(db.Config).Pointer()
	if cached, ok := dbNameCache.Load(cacheKey); ok {
		return cached.(string)
	}

	// 方式3: 通过 SQL 查询获取数据库名（仅在首次访问时执行）
	// 使用 Session 跳过所有插件回调，避免递归
	var dbName string
	sqlDB, err := db.DB()
	if err != nil {
		return "unknown"
	}
	if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&dbName); err != nil || dbName == "" {
		return "unknown"
	}

	// 缓存结果
	dbNameCache.Store(cacheKey, dbName)
	return dbName
}

// extractTableName 从 GORM Statement 提取表名（包含表前缀）
func extractTableName(db *gorm.DB) string {
	// 优先从 Statement.Table 获取（显式指定的表名，已包含前缀）
	if db.Statement.Table != "" {
		return db.Statement.Table
	}

	// 从 Schema 推断（通过 Model 推断，已包含前缀）
	if db.Statement.Schema != nil && db.Statement.Schema.Table != "" {
		return db.Statement.Schema.Table
	}

	// 从 Model 类型推断（需要手动添加前缀）
	if db.Statement.Model != nil {
		modelType := reflect.TypeOf(db.Statement.Model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		if modelType.Kind() == reflect.Slice {
			modelType = modelType.Elem()
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
		}
		// 转换为表名（驼峰转下划线）并添加前缀
		tableName := toSnakeCase(modelType.Name())
		return applyTablePrefix(db, tableName)
	}

	return "unknown"
}

// applyTablePrefix 应用表名前缀
func applyTablePrefix(db *gorm.DB, tableName string) string {
	if db.Config != nil && db.Config.NamingStrategy != nil {
		// 使用 GORM 的 NamingStrategy 获取完整表名
		return db.Config.NamingStrategy.TableName(tableName)
	}
	return tableName
}

// extractTableNameFromMutator 从写操作中提取表名（包含表前缀）
func extractTableNameFromMutator(db *gorm.DB) string {
	// 优先从 Statement.Table 获取（已包含前缀）
	if db.Statement.Table != "" {
		return db.Statement.Table
	}

	// 从 Schema 获取（已包含前缀）
	if db.Statement.Schema != nil && db.Statement.Schema.Table != "" {
		return db.Statement.Schema.Table
	}

	// 从 Model 类型推断（优先于 Dest，因为 Update 操作时 Dest 可能是 map）
	if db.Statement.Model != nil {
		modelType := reflect.TypeOf(db.Statement.Model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		if modelType.Kind() == reflect.Slice {
			modelType = modelType.Elem()
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
		}
		if modelType.Kind() == reflect.Struct {
			tableName := toSnakeCase(modelType.Name())
			return applyTablePrefix(db, tableName)
		}
	}

	// 从 Dest 类型推断（作为 Model 的备选）
	if db.Statement.Dest != nil {
		destType := reflect.TypeOf(db.Statement.Dest)
		if destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		if destType.Kind() == reflect.Slice {
			destType = destType.Elem()
			if destType.Kind() == reflect.Ptr {
				destType = destType.Elem()
			}
		}
		if destType.Kind() == reflect.Struct {
			tableName := toSnakeCase(destType.Name())
			return applyTablePrefix(db, tableName)
		}
	}

	return "unknown"
}

// valuesToString 将参数列表转换为字符串
func valuesToString(vals []any) string {
	if len(vals) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, val := range vals {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(valueToString(val))
	}
	return builder.String()
}

// valueToString 将单个值转换为字符串
func valueToString(val any) string {
	if val == nil {
		return "nil"
	}

	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return "nil"
		}
		return valueToString(v.Elem().Interface())

	case reflect.Map:
		if v.Len() == 0 {
			return "{}"
		}
		var pairs []string
		iter := v.MapRange()
		for iter.Next() {
			pairs = append(pairs, fmt.Sprintf("%v:%v", iter.Key(), iter.Value()))
		}
		return "{" + strings.Join(pairs, ",") + "}"

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "[]"
		}
		var items []string
		for i := 0; i < v.Len(); i++ {
			items = append(items, valueToString(v.Index(i).Interface()))
		}
		return "[" + strings.Join(items, ",") + "]"

	case reflect.Struct:
		// 对于时间等特殊结构体，使用 String() 方法
		if stringer, ok := val.(fmt.Stringer); ok {
			return stringer.String()
		}
		return fmt.Sprintf("%+v", val)

	default:
		return fmt.Sprintf("%v", val)
	}
}

// toSnakeCase 将驼峰命名转换为下划线命名
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
