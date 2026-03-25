package persistence

import (
	goCtx "context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"ttpos-server-go/pkg/context"
)

const (
	// SystemPrefix 系统前缀，用于区分不同系统的缓存 key
	SystemPrefix       = "ttpos5"        // 系统前缀
	CacheVersionPrefix = "cacheversion"  // 缓存版本时间戳前缀
	CacheVersionTTL    = 5 * time.Minute // 缓存版本时间戳的过期时间
	GlobalObjectUuid   = 0               // 全局对象 UUID
)

// BuildKey 构建 key 的辅助方法（自动从 context 提取 company UUID）
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
func BuildKey(ctx goCtx.Context, objectType string, objectUuid uint64) string {
	cctx := ctx.(context.Context)
	companyUuid := cctx.GetCompanyUuid()
	return BuildKeyWithCompanyUuid(companyUuid, objectType, objectUuid)
}

func BuildKeyWithCompanyUuid(companyUuid uint64, objectType string, objectUuid uint64) string {
	return fmt.Sprintf("%s:%d:%s:%d", SystemPrefix, companyUuid, objectType, objectUuid)
}

// BuildAuthStaffKey 构建员工信息缓存 key
// Key 格式：{system_prefix}:{company_uuid}:staff:{staff_uuid}
func BuildAuthStaffKey(ctx goCtx.Context, staffUuid uint64) string {
	return BuildKey(ctx, ObjectTypeStaff, staffUuid)
}

// BuildApiPermissionKey 构建 API 权限缓存 key
// Key 格式：{system_prefix}:{company_uuid}:api_permission:{staff_uuid}
func BuildApiPermissionKey(companyUuid, staffUuid uint64) string {
	return BuildKeyWithCompanyUuid(companyUuid, ObjectTypeApiPermission, staffUuid)
}

// BuildProductListCacheKey 构建商品列表缓存 key
// Key 格式：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{orderType}:{recommendUuids}
// 参数：
//   - companyUuid: 门店 UUID
//   - source: 来源（如 cashier, assistant 等）
//   - pageNo: 页码
//   - pageSize: 每页数量
//   - isMember: 是否会员
//   - orderType: 会员端订单类型（0-外送，1-堂食/到店自取）
//   - recommendUuids: 推荐商品 UUID 列表（会自动排序以确保相同 UUIDs 但顺序不同时 key 一致）
func BuildProductListCacheKey(companyUuid uint64, source string, pageNo, pageSize int, isMember bool, orderType int, recommendUuids []uint64) string {
	// 对推荐商品 UUIDs 排序，确保相同 UUIDs 但顺序不同时 key 一致
	sortedUuids := make([]uint64, len(recommendUuids))
	copy(sortedUuids, recommendUuids)
	sort.Slice(sortedUuids, func(i, j int) bool {
		return sortedUuids[i] < sortedUuids[j]
	})

	// 构建 key：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{orderType}:{recommendUuids}
	keyParts := []string{
		SystemPrefix,
		fmt.Sprintf("%d", companyUuid),
		ObjectTypeProductList,
		source,
		fmt.Sprintf("%d", pageNo),
		fmt.Sprintf("%d", pageSize),
		fmt.Sprintf("%v", isMember),
		fmt.Sprintf("%d", orderType),
	}

	// 添加推荐商品 UUIDs
	if len(sortedUuids) > 0 {
		uuidStrs := make([]string, len(sortedUuids))
		for i, uuid := range sortedUuids {
			uuidStrs[i] = fmt.Sprintf("%d", uuid)
		}
		keyParts = append(keyParts, strings.Join(uuidStrs, ","))
	}

	return strings.Join(keyParts, ":")
}

// BuildCacheVersionKey 构建缓存版本时间戳 key
// Key 格式：{cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
// 注意：object_uuid 通常为 0，表示全局版本时间戳
// 参数：
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型
//   - objectUuid: 对象 UUID（通常为 0，表示全局版本时间戳）
//
// 返回：
//   - string: 缓存版本时间戳 key
func BuildCacheVersionKey(companyUuid uint64, objectType string, objectUuid uint64) string {
	return fmt.Sprintf("%s:%s:%d:%s:%d", CacheVersionPrefix, SystemPrefix, companyUuid, objectType, objectUuid)
}

// GetUpdateTime 使用反射获取对象的 UpdateTime 字段值
// 参数：
//   - obj: 对象实例（可以是值类型、指针类型或指向指针的指针）
//
// 返回：
//   - int64: UpdateTime 字段值，如果字段不存在或无法获取则返回 0
func GetUpdateTime(obj interface{}) int64 {
	val := reflect.ValueOf(obj)
	// 如果是指针，需要解引用到实际的结构体
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}

	// 确保是结构体类型
	if val.Kind() != reflect.Struct {
		return 0
	}

	// 查找 UpdateTime 字段
	field := val.FieldByName("UpdateTime")
	if !field.IsValid() || !field.CanInterface() {
		return 0
	}

	// 转换为 int64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(field.Uint())
	default:
		return 0
	}
}

// SetUpdateTime 使用反射设置对象的 UpdateTime 字段值
// 参数：
//   - obj: 对象实例指针（可以是 *T 或 **T，函数会自动处理）
//   - updateTime: 要设置的更新时间戳（秒）
func SetUpdateTime(obj interface{}, updateTime int64) {
	val := reflect.ValueOf(obj)
	// 如果是指针，需要解引用到实际的结构体
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	// 确保是结构体类型
	if val.Kind() != reflect.Struct {
		return
	}

	// 查找 UpdateTime 字段
	field := val.FieldByName("UpdateTime")
	if !field.IsValid() || !field.CanSet() {
		return
	}

	// 设置值
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(updateTime)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(updateTime))
	}
}
