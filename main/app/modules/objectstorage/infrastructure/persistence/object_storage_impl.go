package persistence

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"ttpos-server-go/app/modules/objectstorage/domain/entity"
	"ttpos-server-go/app/modules/objectstorage/domain/service"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ObjectStorageImpl 对象存储实现
type ObjectStorageImpl[T any] struct {
	config *service.Config
}

// NewObjectStorage 创建对象存储实例
func NewObjectStorage[T any](config *service.Config) service.IObjectStorage[T] {
	return &ObjectStorageImpl[T]{
		config: config,
	}
}

// Get 获取对象
func (s *ObjectStorageImpl[T]) Get(ctx context.Context, key string, query func() (T, error)) (T, error) {
	var zero T

	// 如果禁用缓存，直接调用查询函数
	if s.config.DisableCache {
		return query()
	}

	// 从三级缓存获取
	result, err := s.config.CacheLayer.GET(key, func() (any, error) {
		return query()
	})

	if err != nil {
		return zero, err
	}

	// 类型断言
	if typed, ok := result.(T); ok {
		return typed, nil
	}

	return zero, fmt.Errorf("类型断言失败: 期望 %T，实际 %T", zero, result)
}

// BatchGet 批量获取对象
func (s *ObjectStorageImpl[T]) BatchGet(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) (map[string]T, error) {
	// 去重
	uniqueKeys := deduplicate(keys)

	// 如果禁用缓存，直接调用查询函数
	if s.config.DisableCache {
		return query(uniqueKeys)
	}

	// 批量从三级缓存获取
	result, err := s.config.CacheLayer.BATCH_GET(uniqueKeys, func(missedKeys []string) (map[string]any, error) {
		typedResult, err := query(missedKeys)
		if err != nil {
			return nil, err
		}

		// 转换为 any 类型
		anyResult := make(map[string]any)
		for k, v := range typedResult {
			anyResult[k] = v
		}
		return anyResult, nil
	})

	if err != nil {
		return nil, err
	}

	// 类型转换
	typedResult := make(map[string]T)
	for k, v := range result {
		if typed, ok := v.(T); ok {
			typedResult[k] = typed
		}
	}

	return typedResult, nil
}

// Invalidate 使缓存失效
func (s *ObjectStorageImpl[T]) Invalidate(ctx context.Context, key string) error {
	s.config.CacheLayer.DEL(key)
	return nil
}

// Update 更新缓存
func (s *ObjectStorageImpl[T]) Update(ctx context.Context, key string, value T) error {
	objectType := extractObjectTypeFromKey(key)
	ttl := s.config.GetTTL(objectType)
	return s.config.CacheLayer.SET(key, value, ttl)
}

// Warmup 预热缓存
func (s *ObjectStorageImpl[T]) Warmup(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) error {
	result, err := query(keys)
	if err != nil {
		return err
	}

	// 批量写入缓存
	for key, value := range result {
		objectType := extractObjectTypeFromKey(key)
		ttl := s.config.GetTTL(objectType)
		if err := s.config.CacheLayer.SET(key, value, ttl); err != nil {
			// 记录错误但继续处理其他 key
			continue
		}
	}
	return nil
}

// InvalidateByCompany 按 company 粒度批量失效缓存
func (s *ObjectStorageImpl[T]) InvalidateByCompany(ctx context.Context, companyUuid uint64) error {
	pattern := fmt.Sprintf("%d:*", companyUuid)
	return s.invalidateByPattern(ctx, pattern)
}

// InvalidateByCompanyAndType 按 company + object_type 粒度批量失效缓存
func (s *ObjectStorageImpl[T]) InvalidateByCompanyAndType(ctx context.Context, companyUuid uint64, objectType string) error {
	pattern := fmt.Sprintf("%d:%s:*", companyUuid, objectType)
	return s.invalidateByPattern(ctx, pattern)
}

// invalidateByPattern 根据模式批量失效缓存（内部方法）
func (s *ObjectStorageImpl[T]) invalidateByPattern(ctx context.Context, pattern string) error {
	keys, err := s.config.CacheLayer.SCAN(ctx, pattern)
	if err != nil {
		return fmt.Errorf("扫描缓存 key 失败: %w", err)
	}
	if len(keys) > 0 {
		s.config.CacheLayer.DEL(keys...)
	}
	return nil
}

// UpdateByCompany 按 company 粒度批量更新缓存
func (s *ObjectStorageImpl[T]) UpdateByCompany(ctx context.Context, companyUuid uint64, objectType string, values map[string]T) error {
	ttl := s.config.GetTTL(objectType)
	var lastErr error
	for key, value := range values {
		if err := s.config.CacheLayer.SET(key, value, ttl); err != nil {
			lastErr = err
			continue
		}
	}
	return lastErr
}

// PreloadWithConfig 配置映射自动关联注入的主方法
func (s *ObjectStorageImpl[T]) PreloadWithConfig(ctx context.Context, obj interface{}, associations []entity.Association) error {
	if obj == nil {
		return fmt.Errorf("对象不能为 nil")
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			return fmt.Errorf("对象指针不能为 nil")
		}
		objValue = objValue.Elem()
	}

	// 处理每个关联配置
	for _, assoc := range associations {
		if err := s.preloadAssociation(ctx, objValue, assoc); err != nil {
			logger.Logger.Warn("关联注入失败", zap.String("path", assoc.Path), zap.Error(err))
			continue
		}
	}

	return nil
}

// preloadAssociation 处理单个关联配置的注入
func (s *ObjectStorageImpl[T]) preloadAssociation(ctx context.Context, objValue reflect.Value, assoc entity.Association) error {
	pathParts := parsePath(assoc.Path)
	if len(pathParts) == 0 {
		return fmt.Errorf("路径为空")
	}
	return s.preloadNestedPath(ctx, objValue, pathParts, assoc, 0)
}

// parsePath 解析嵌套路径
func parsePath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// findField 通过反射查找结构体字段
func findField(obj reflect.Value, fieldName string) (reflect.Value, bool) {
	if obj.Kind() == reflect.Ptr {
		if obj.IsNil() {
			return reflect.Value{}, false
		}
		obj = obj.Elem()
	}
	if obj.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	field := obj.FieldByName(fieldName)
	if !field.IsValid() {
		return reflect.Value{}, false
	}
	return field, true
}

// extractUUID 调用 GetUUID 函数从对象中提取 UUID
func extractUUID(obj interface{}, getUUID func(obj interface{}) uint64) uint64 {
	if getUUID == nil {
		return 0
	}
	return getUUID(obj)
}

// collectUUIDs 收集同一层级的 UUID，用于批量查询
func collectUUIDs(objs reflect.Value, getUUID func(obj interface{}) uint64) []uint64 {
	if objs.Kind() != reflect.Slice {
		return nil
	}
	seen := make(map[uint64]bool)
	var uuids []uint64
	for i := 0; i < objs.Len(); i++ {
		item := objs.Index(i)
		uuid := extractUUID(item.Interface(), getUUID)
		if uuid != 0 && !seen[uuid] {
			seen[uuid] = true
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

// batchQuery 调用 BatchQueryFunc 批量查询对象
func batchQuery(ctx context.Context, assoc entity.Association, uuids []uint64) (map[uint64]interface{}, error) {
	if len(uuids) == 0 {
		return make(map[uint64]interface{}), nil
	}
	if assoc.BatchQueryFunc != nil {
		return assoc.BatchQueryFunc(ctx, uuids)
	}
	result := make(map[uint64]interface{})
	for _, uuid := range uuids {
		if assoc.QueryFunc != nil {
			obj, err := assoc.QueryFunc(ctx, uuid)
			if err != nil {
				logger.Logger.Warn("批量查询单个对象失败", zap.Uint64("uuid", uuid), zap.Error(err))
				continue
			}
			result[uuid] = obj
		}
	}
	return result, nil
}

// setField 使用反射设置结构体字段
func setField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("字段不可设置")
	}
	if value == nil {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}
	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return fmt.Errorf("值无效")
	}
	fieldType := field.Type()
	valType := val.Type()
	if fieldType.Kind() == reflect.Ptr {
		if valType.Kind() == reflect.Ptr {
			if valType.AssignableTo(fieldType) {
				field.Set(val)
			} else if valType.Elem().AssignableTo(fieldType.Elem()) {
				ptr := reflect.New(fieldType.Elem())
				ptr.Elem().Set(val.Elem())
				field.Set(ptr)
			} else {
				return fmt.Errorf("类型不匹配: 字段类型 %v, 值类型 %v", fieldType, valType)
			}
		} else {
			if valType.AssignableTo(fieldType.Elem()) {
				ptr := reflect.New(fieldType.Elem())
				ptr.Elem().Set(val)
				field.Set(ptr)
			} else {
				return fmt.Errorf("类型不匹配: 字段类型 %v, 值类型 %v", fieldType.Elem(), valType)
			}
		}
	} else {
		if valType.Kind() == reflect.Ptr {
			if val.IsNil() {
				return fmt.Errorf("值为 nil")
			}
			if valType.Elem().AssignableTo(fieldType) {
				field.Set(val.Elem())
			} else {
				return fmt.Errorf("类型不匹配: 字段类型 %v, 值类型 %v", fieldType, valType.Elem())
			}
		} else {
			if valType.AssignableTo(fieldType) {
				field.Set(val)
			} else {
				return fmt.Errorf("类型不匹配: 字段类型 %v, 值类型 %v", fieldType, valType)
			}
		}
	}
	return nil
}

// preloadNestedPath 递归处理嵌套路径
func (s *ObjectStorageImpl[T]) preloadNestedPath(ctx context.Context, parentValue reflect.Value, pathParts []string, assoc entity.Association, depth int) error {
	if depth >= len(pathParts) {
		return nil
	}
	currentFieldName := pathParts[depth]
	isLast := depth == len(pathParts)-1
	fieldValue, found := findField(parentValue, currentFieldName)
	if !found {
		return fmt.Errorf("字段 %s 不存在", currentFieldName)
	}
	if isLast {
		return s.injectField(ctx, parentValue, fieldValue, assoc)
	}
	if fieldValue.Kind() == reflect.Slice {
		for i := 0; i < fieldValue.Len(); i++ {
			item := fieldValue.Index(i)
			if err := s.preloadNestedPath(ctx, item, pathParts, assoc, depth+1); err != nil {
				logger.Logger.Warn("嵌套路径注入失败", zap.String("field", currentFieldName), zap.Int("index", i), zap.Error(err))
				continue
			}
		}
		return nil
	}
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil
		}
		return s.preloadNestedPath(ctx, fieldValue.Elem(), pathParts, assoc, depth+1)
	}
	return s.preloadNestedPath(ctx, fieldValue, pathParts, assoc, depth+1)
}

// injectField 注入字段值
func (s *ObjectStorageImpl[T]) injectField(ctx context.Context, parentValue reflect.Value, fieldValue reflect.Value, assoc entity.Association) error {
	pathParts := parsePath(assoc.Path)
	var assocFieldName string
	if len(pathParts) > 0 {
		assocFieldName = pathParts[len(pathParts)-1]
	}
	if fieldValue.Kind() == reflect.Slice {
		if fieldValue.Len() == 0 {
			return nil
		}
		uuids := collectUUIDs(fieldValue, assoc.GetUUID)
		if len(uuids) == 0 {
			return nil
		}
		results, err := batchQuery(ctx, assoc, uuids)
		if err != nil {
			return fmt.Errorf("批量查询失败: %w", err)
		}
		for i := 0; i < fieldValue.Len(); i++ {
			item := fieldValue.Index(i)
			var itemValue reflect.Value
			if item.Kind() == reflect.Ptr {
				if item.IsNil() {
					continue
				}
				itemValue = item.Elem()
			} else {
				itemValue = item
			}
			uuid := extractUUID(itemValue.Interface(), assoc.GetUUID)
			if uuid != 0 {
				if result, ok := results[uuid]; ok {
					var targetValue reflect.Value
					if item.Kind() == reflect.Ptr {
						targetValue = item.Elem()
					} else {
						targetValue = item
					}
					if assocField, found := findField(targetValue, assocFieldName); found {
						if err := setField(assocField, result); err != nil {
							logger.Logger.Warn("设置关联字段失败", zap.String("field", assocFieldName), zap.Uint64("uuid", uuid), zap.Error(err))
						}
					}
				}
			}
		}
		return nil
	}
	uuid := extractUUID(parentValue.Interface(), assoc.GetUUID)
	if uuid == 0 {
		return nil
	}
	var result interface{}
	var err error
	if assoc.BatchQueryFunc != nil {
		results, err := assoc.BatchQueryFunc(ctx, []uint64{uuid})
		if err == nil {
			result = results[uuid]
		}
	} else if assoc.QueryFunc != nil {
		result, err = assoc.QueryFunc(ctx, uuid)
	}
	if err != nil {
		return fmt.Errorf("查询对象失败: %w", err)
	}
	if result == nil {
		return nil
	}
	return setField(fieldValue, result)
}
