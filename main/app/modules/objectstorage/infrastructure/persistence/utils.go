package persistence

import (
	goCtx "context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// SystemPrefix 系统前缀，用于区分不同系统的缓存 key
	SystemPrefix = "ttpos5"
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
// Key 格式：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{recommendUuids}
// 参数：
//   - companyUuid: 门店 UUID
//   - source: 来源（如 cashier, assistant 等）
//   - pageNo: 页码
//   - pageSize: 每页数量
//   - isMember: 是否会员
//   - recommendUuids: 推荐商品 UUID 列表（会自动排序以确保相同 UUIDs 但顺序不同时 key 一致）
func BuildProductListCacheKey(companyUuid uint64, source string, pageNo, pageSize int, isMember bool, recommendUuids []uint64) string {
	// 对推荐商品 UUIDs 排序，确保相同 UUIDs 但顺序不同时 key 一致
	sortedUuids := make([]uint64, len(recommendUuids))
	copy(sortedUuids, recommendUuids)
	sort.Slice(sortedUuids, func(i, j int) bool {
		return sortedUuids[i] < sortedUuids[j]
	})

	// 构建 key：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{recommendUuids}
	keyParts := []string{
		SystemPrefix,
		fmt.Sprintf("%d", companyUuid),
		ObjectTypeProductList,
		source,
		fmt.Sprintf("%d", pageNo),
		fmt.Sprintf("%d", pageSize),
		fmt.Sprintf("%v", isMember),
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
// Key 格式：{system_prefix}:{company_uuid}:cacheversion_{object_type}
// 参数：
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型（如 "product_list"）
//
// 返回：
//   - string: 缓存版本时间戳 key
func BuildCacheVersionKey(companyUuid uint64, objectType string) string {
	return fmt.Sprintf("%s:%d:cacheversion_%s", SystemPrefix, companyUuid, objectType)
}

// ExtractCompanyUuidFromCacheKey 从缓存 key 中提取公司 UUID
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:...
// 例如：ttpos5:123456:product_list:cashier:1:20:false
// 参数：
//   - cacheKey: 缓存 key 字符串
//
// 返回：
//   - uint64: 公司 UUID
//   - error: 错误信息
func ExtractCompanyUuidFromCacheKey(cacheKey string) (uint64, error) {
	// Key 格式：{system_prefix}:{company_uuid}:{object_type}:...
	// 公司 UUID 是 key 的第二个部分（索引为 1）
	parts := strings.Split(cacheKey, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("无效的缓存 key 格式: %s", cacheKey)
	}
	companyUuidStr := parts[1]
	companyUuid, err := strconv.ParseUint(companyUuidStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析公司 UUID: %v", err)
	}
	return companyUuid, nil
}

// GetCacheVersionTimestamp 获取缓存版本时间戳
// 参数：
//   - cacheInstance: 缓存实例（用于获取 Redis 客户端）
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型
//
// 返回：
//   - int64: 版本时间戳（Unix 时间戳，秒）
//   - bool: 是否存在版本时间戳
func GetCacheVersionTimestamp(cacheInstance cache.Cache, companyUuid uint64, objectType string) (int64, bool) {
	// 获取 Redis 客户端
	var client redis.UniversalClient
	if clusterClient := cacheInstance.GetClusterClient(); clusterClient != nil {
		client = clusterClient
	} else if redisClient := cacheInstance.GetClient(); redisClient != nil {
		client = redisClient
	} else {
		// 如果不是 Redis 缓存，返回 false
		return 0, false
	}

	// 构建版本时间戳 key
	versionKey := BuildCacheVersionKey(companyUuid, objectType)

	// 从 Redis 获取版本时间戳
	backgroundCtx := goCtx.Background()
	timestampStr, err := client.Get(backgroundCtx, versionKey).Result()
	if err == redis.Nil {
		// 不存在版本时间戳，返回 false
		return 0, false
	}
	if err != nil {
		// 获取失败，返回 false
		return 0, false
	}

	// 解析时间戳
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, false
	}

	logger.Logger.Debug("CacheVersionTimestamp query", zap.Int64("timestamp", timestamp), zap.String("versionKey", versionKey))
	return timestamp, true
}

// UpdateCacheVersionTimestamp 更新缓存版本时间戳
// 参数：
//   - cacheInstance: 缓存实例（用于获取 Redis 客户端）
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型
//   - ttl: 版本时间戳的过期时间（应该设置为最长有效的缓存时间，如 L2TTL）
//     当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
//     表示缓存已过期，需要重新查询并设置新的版本时间戳
//
// 返回：
//   - error: 错误信息
func UpdateCacheVersionTimestamp(cacheInstance cache.Cache, companyUuid uint64, objectType string, ttl time.Duration) error {
	// 获取 Redis 客户端
	var client redis.UniversalClient
	if clusterClient := cacheInstance.GetClusterClient(); clusterClient != nil {
		client = clusterClient
	} else if redisClient := cacheInstance.GetClient(); redisClient != nil {
		client = redisClient
	} else {
		// 如果不是 Redis 缓存，无法设置版本时间戳，直接返回
		return nil
	}

	// 构建版本时间戳 key
	versionKey := BuildCacheVersionKey(companyUuid, objectType)

	// 使用当前时间戳（秒）作为版本时间戳
	currentTimestamp := time.Now().Unix()

	// 设置版本时间戳，过期时间应该和最长有效的缓存时间一致（如 L2TTL）
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	backgroundCtx := goCtx.Background()
	if err := client.Set(backgroundCtx, versionKey, currentTimestamp, ttl).Err(); err != nil {
		return fmt.Errorf("设置缓存版本时间戳失败: %w", err)
	}

	logger.Logger.Debug("CacheVersionTimestamp update", zap.Int64("currentTimestamp", currentTimestamp), zap.String("versionKey", versionKey), zap.Duration("ttl", ttl))
	return nil
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
