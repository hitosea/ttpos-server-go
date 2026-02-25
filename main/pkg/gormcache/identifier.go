package gormcache

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

const (
	// IdentifierPrefix 缓存键前缀
	IdentifierPrefix = "gorm:cache:"
)

// buildIdentifier 构建缓存键
// 格式: gorm:cache:{tableName}:{sql}:{params}
func buildIdentifier(db *gorm.DB) (tableName string, key string) {
	tableName = extractTableName(db)

	// 触发 SQL 编译（不执行）
	db.Statement.Build("SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR")

	// 构建唯一标识
	var builder strings.Builder
	builder.WriteString(IdentifierPrefix)
	builder.WriteString(tableName)
	builder.WriteString(":")
	builder.WriteString(db.Statement.SQL.String())
	builder.WriteString(":")
	builder.WriteString(valuesToString(db.Statement.Vars))

	key = builder.String()
	return tableName, key
}

// extractTableName 从 GORM Statement 提取表名
func extractTableName(db *gorm.DB) string {
	// 优先从 Statement.Table 获取（显式指定的表名）
	if db.Statement.Table != "" {
		return db.Statement.Table
	}

	// 从 Schema 推断（通过 Model 推断）
	if db.Statement.Schema != nil && db.Statement.Schema.Table != "" {
		return db.Statement.Schema.Table
	}

	// 从 Model 类型推断
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
		// 转换为表名（驼峰转下划线）
		return toSnakeCase(modelType.Name())
	}

	return "unknown"
}

// extractTableNameFromMutator 从写操作中提取表名
func extractTableNameFromMutator(db *gorm.DB) string {
	// 优先从 Statement.Table 获取
	if db.Statement.Table != "" {
		return db.Statement.Table
	}

	// 从 Schema 获取
	if db.Statement.Schema != nil && db.Statement.Schema.Table != "" {
		return db.Statement.Schema.Table
	}

	// 从 Dest 类型推断
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
			return toSnakeCase(destType.Name())
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
