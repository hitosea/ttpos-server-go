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
}

// NewObjectStorage 创建对象存储实例
func NewObjectStorage[T any]() service.IObjectStorage[T] {
	return &ObjectStorageImpl[T]{}
}

// PreloadWithConfig 配置映射自动关联注入的主方法
//
// 功能说明：
//
//	类似于 GORM 的 Preload 功能，通过配置化的方式自动注入对象的关联数据。
//	支持嵌套路径（如 "SaleOrders.SaleOrderProducts.ProductPackage"）和批量查询优化。
//
// 参数说明：
//   - ctx: 上下文对象，用于传递公司UUID等上下文信息
//   - obj: 待注入关联对象的对象实例（支持指针类型和非指针类型）
//   - associations: 关联配置列表，每个配置定义了一个关联路径和查询方式
//
// 返回值：
//   - error: 如果所有关联注入都失败则返回错误，单个关联失败只记录警告日志
//
// 使用示例：
//
//	associations := []entity.Association{
//	    {
//	        Path:       "Desk",
//	        ObjectType: "desk",
//	        GetUUID: func(obj interface{}) uint64 {
//	            return obj.(model.SaleBill).DeskUuid
//	        },
//	        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
//	            return deskRepo.GetDesk(CommonRepo.WhereByUuid(uuid))
//	        },
//	    },
//	}
//	err := objectStorage.PreloadWithConfig(ctx, &saleBill, associations)
//
// 实现逻辑：
//  1. 参数校验：检查对象是否为 nil
//  2. 反射处理：将对象转换为 reflect.Value，处理指针类型
//  3. 遍历配置：对每个关联配置调用 preloadAssociation 进行注入
//  4. 错误处理：单个关联失败不影响其他关联的注入，只记录警告日志
func (s *ObjectStorageImpl[T]) PreloadWithConfig(ctx context.Context, obj interface{}, associations []entity.Association) error {
	// 参数校验：检查对象是否为 nil
	if obj == nil {
		return fmt.Errorf("对象不能为 nil")
	}

	// 使用反射获取对象的值
	// 支持指针类型和非指针类型，统一转换为非指针的 reflect.Value
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		// 如果是指针类型，检查是否为 nil
		if objValue.IsNil() {
			return fmt.Errorf("对象指针不能为 nil")
		}
		// 获取指针指向的实际值
		objValue = objValue.Elem()
	}

	// 遍历每个关联配置，依次进行注入
	// 注意：单个关联注入失败不会中断整个流程，只记录警告日志
	for _, assoc := range associations {
		if err := s.preloadAssociation(ctx, objValue, assoc); err != nil {
			// 关联注入失败时记录警告日志，但继续处理下一个关联配置
			// 这样可以保证部分关联注入成功时，对象仍然可用
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

// preloadNestedPath 递归处理嵌套路径的关联对象注入
//
// 功能说明：
//
//	该方法递归遍历嵌套路径（如 "BuffetPackage1.MultiLanguageName"），逐层定位字段并最终注入关联对象。
//	支持处理切片类型（一对多关系）和指针类型（一对一关系），确保能够正确访问嵌套结构。
//
// 参数说明：
//   - ctx: 上下文对象，用于传递公司UUID等上下文信息
//   - parentValue: 当前层级的父对象反射值（可能是结构体、指针或切片元素）
//   - pathParts: 路径分割后的字段名数组，如 ["BuffetPackage1", "MultiLanguageName"]
//   - assoc: 关联配置，包含查询函数和UUID提取函数
//   - depth: 当前递归深度，从 0 开始，对应 pathParts 的索引
//
// 返回值：
//   - error: 如果字段不存在或注入失败则返回错误；nil 指针或空切片会静默跳过
//
// 实现逻辑：
//  1. 边界检查：如果深度超出路径长度，直接返回（递归终止条件）
//  2. 字段查找：根据当前深度对应的字段名，在父对象中查找字段
//  3. 最后字段判断：如果是路径的最后一个字段，调用 injectField 进行注入
//  4. 切片处理：如果当前字段是切片类型，遍历每个元素并递归处理
//  5. 指针处理：如果当前字段是指针类型，先解引用再递归处理
//  6. 普通字段：直接递归处理下一层路径
//
// 使用示例：
//
//	路径 "BuffetPackage1.MultiLanguageName" 的处理流程：
//	- depth=0: 查找 SaleBill.BuffetPackage1（指针类型）→ 解引用 → 递归
//	- depth=1: 查找 BuffetPackage.MultiLanguageName（最后字段）→ 调用 injectField 注入
//
//	路径 "BuffetPackage1.BuffetProducts.ProductPackage" 的处理流程：
//	- depth=0: 查找 SaleBill.BuffetPackage1（指针类型）→ 解引用 → 递归
//	- depth=1: 查找 BuffetPackage.BuffetProducts（切片类型）→ 遍历每个元素 → 递归
//	- depth=2: 查找 BuffetProduct.ProductPackage（最后字段）→ 调用 injectField 注入
func (s *ObjectStorageImpl[T]) preloadNestedPath(ctx context.Context, parentValue reflect.Value, pathParts []string, assoc entity.Association, depth int) error {
	// 边界检查：递归终止条件
	if depth >= len(pathParts) {
		return nil
	}

	// 获取当前深度的字段名
	currentFieldName := pathParts[depth]
	isLast := depth == len(pathParts)-1

	// 在当前父对象中查找字段
	fieldValue, found := findField(parentValue, currentFieldName)
	if !found {
		return fmt.Errorf("字段 %s 不存在", currentFieldName)
	}

	// 如果是路径的最后一个字段，直接进行注入
	if isLast {
		return s.injectField(ctx, parentValue, fieldValue, assoc)
	}

	// 处理切片类型（一对多关系）：遍历每个元素并递归处理
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

	// 处理指针类型（一对一关系）：先解引用再递归处理
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			// nil 指针静默跳过，不进行注入
			return nil
		}
		return s.preloadNestedPath(ctx, fieldValue.Elem(), pathParts, assoc, depth+1)
	}

	// 普通字段：直接递归处理下一层路径
	return s.preloadNestedPath(ctx, fieldValue, pathParts, assoc, depth+1)
}

// injectField 注入关联对象到目标字段
//
// 功能说明：
//
//	该方法负责将查询到的关联对象注入到目标字段中。支持三种场景：
//	1. 直接注入列表到切片字段（如 BuffetPackage1.BuffetProducts）：从父对象提取 UUID，查询整个列表，直接注入到切片字段
//	2. 注入到切片元素的嵌套字段（如 BuffetProducts.ProductPackage）：遍历切片元素，批量查询关联对象，注入到每个元素的嵌套字段中
//	3. 单个字段（一对一关系）：从父对象提取 UUID，查询关联对象，直接注入到字段中
//
// 参数说明：
//   - ctx: 上下文对象，用于传递公司UUID等上下文信息
//   - parentValue: 父对象的反射值，用于提取 UUID（一对一关系场景）
//   - fieldValue: 目标字段的反射值，可能是切片类型或单个字段类型
//   - assoc: 关联配置，包含查询函数、批量查询函数和UUID提取函数
//
// 返回值：
//   - error: 如果查询失败或注入失败则返回错误；空切片或 UUID 为 0 会静默跳过
//
// 实现逻辑：
//
//	场景 1: 直接注入列表到切片字段（如 BuffetPackage1.BuffetProducts）
//	1. 判断路径是否为两层（如 "BuffetPackage1.BuffetProducts"）
//	2. 从父对象提取 UUID（通过 GetUUID 函数）
//	3. 如果 UUID 为 0，尝试场景 2 的处理逻辑
//	4. 查询关联对象列表（优先使用 BatchQueryFunc，否则使用 QueryFunc）
//	5. 直接将查询结果（列表）注入到切片字段中
//
//	场景 2: 注入到切片元素的嵌套字段（如 BuffetProducts.ProductPackage）
//	1. 检查切片是否为空，为空则直接返回
//	2. 收集切片中所有元素的 UUID（通过 GetUUID 函数）
//	3. 批量查询所有关联对象（优先使用 BatchQueryFunc，否则使用 QueryFunc）
//	4. 遍历切片元素，根据 UUID 匹配查询结果，注入到元素的嵌套字段中
//
//	场景 3: 单个字段处理流程（一对一关系，如 Desk、MultiLanguageName）
//	1. 从父对象提取 UUID（通过 GetUUID 函数）
//	2. 如果 UUID 为 0，直接返回（无需查询）
//	3. 查询关联对象（优先使用 BatchQueryFunc，否则使用 QueryFunc）
//	4. 将查询结果注入到目标字段中
//
// 使用示例：
//
//	场景 1: 直接注入列表到切片字段（一对多关系）
//	路径 "BuffetPackage1.BuffetProducts"，parentValue 是 BuffetPackage1，fieldValue 是 BuffetProducts 切片字段
//	- 从 BuffetPackage1 提取 Uuid
//	- 查询整个 BuffetProducts 列表
//	- 直接将列表注入到 BuffetPackage1.BuffetProducts 字段中
//
//	场景 2: 注入到切片元素的嵌套字段（一对多关系）
//	路径 "BuffetProducts.ProductPackage"，parentValue 是 BuffetPackage，fieldValue 是 BuffetProducts 切片
//	- 遍历 BuffetProducts 切片，收集每个 BuffetProduct 的 ProductPackageUuid
//	- 批量查询所有 ProductPackage 对象
//	- 将每个 ProductPackage 注入到对应 BuffetProduct.ProductPackage 字段中
//
//	场景 3: 注入单个对象（一对一关系）
//	路径 "Desk"，parentValue 是 SaleBill，fieldValue 是 SaleBill.Desk 字段
//	- 从 SaleBill 提取 DeskUuid
//	- 查询 Desk 对象
//	- 注入到 SaleBill.Desk 字段
func (s *ObjectStorageImpl[T]) injectField(ctx context.Context, parentValue reflect.Value, fieldValue reflect.Value, assoc entity.Association) error {
	// 解析路径，获取最后一个字段名（用于嵌套注入场景）
	pathParts := parsePath(assoc.Path)
	var assocFieldName string
	if len(pathParts) > 0 {
		assocFieldName = pathParts[len(pathParts)-1]
	}

	// 处理切片类型（一对多关系）
	if fieldValue.Kind() == reflect.Slice {
		// 场景 1: 直接注入列表到切片字段（如 BuffetPackage1.BuffetProducts）
		// 判断条件：路径只有两层，且可以从父对象中提取 UUID
		if len(pathParts) == 2 {
			// 尝试从父对象提取 UUID
			uuid := extractUUID(parentValue.Interface(), assoc.GetUUID)
			if uuid != 0 {
				// 查询关联对象列表（优先使用批量查询函数，否则使用单个查询函数）
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
					return fmt.Errorf("查询对象列表失败: %w", err)
				}
				if result == nil {
					return nil
				}

				// 直接将查询结果（列表）注入到切片字段中
				return setField(fieldValue, result)
			}
		}

		// 场景 2: 注入到切片元素的嵌套字段（如 BuffetProducts.ProductPackage）
		// 空切片直接返回
		if fieldValue.Len() == 0 {
			return nil
		}

		// 收集切片中所有元素的 UUID
		uuids := collectUUIDs(fieldValue, assoc.GetUUID)
		if len(uuids) == 0 {
			return nil
		}

		// 批量查询所有关联对象
		results, err := batchQuery(ctx, assoc, uuids)
		if err != nil {
			return fmt.Errorf("批量查询失败: %w", err)
		}

		// 遍历切片元素，将查询结果注入到每个元素的嵌套字段中
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

			// 从切片元素中提取 UUID
			uuid := extractUUID(itemValue.Interface(), assoc.GetUUID)
			if uuid != 0 {
				if result, ok := results[uuid]; ok {
					// 确定目标值（可能是指针或值类型）
					var targetValue reflect.Value
					if item.Kind() == reflect.Ptr {
						targetValue = item.Elem()
					} else {
						targetValue = item
					}

					// 查找嵌套字段并注入关联对象
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

	// 处理单个字段（一对一关系）：从父对象提取 UUID，查询并注入
	uuid := extractUUID(parentValue.Interface(), assoc.GetUUID)
	if uuid == 0 {
		return nil
	}

	// 查询关联对象（优先使用批量查询函数，否则使用单个查询函数）
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

	// 将查询结果注入到目标字段中
	return setField(fieldValue, result)
}
