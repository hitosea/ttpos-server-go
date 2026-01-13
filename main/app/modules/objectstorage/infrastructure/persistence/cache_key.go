package persistence

import (
	"fmt"
	"strconv"
	"strings"
)

// CacheKey 缓存 key 结构体
// 用于解析和构建缓存 key 的各个部分
type CacheKey struct {
	SystemPrefix string   // 系统前缀，如 "ttpos5"
	CompanyUuid  uint64   // 公司 UUID
	ObjectType   string   // 对象类型，如 "product_list", "staff" 等
	Parts        []string // 其他部分（可变部分，如 source, pageNo, pageSize, object_uuid 等）
	IsVersionKey bool     // 是否是版本时间戳 key（以 cacheversion 开头）
}

// Parse 解析缓存 key 字符串
// 支持的格式：
//   - 普通对象 key: {system_prefix}:{company_uuid}:{object_type}:{object_uuid}
//   - 商品列表 key: {system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{recommendUuids}
//   - 版本时间戳 key: {cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
//
// 参数：
//   - key: 缓存 key 字符串
//
// 返回：
//   - error: 错误信息
func (ck *CacheKey) Parse(key string) error {
	parts := strings.Split(key, ":")
	if len(parts) < 3 {
		return fmt.Errorf("无效的缓存 key 格式: %s", key)
	}

	// 检查是否是版本时间戳 key（以 cacheversion 开头）
	if parts[0] == CacheVersionPrefix {
		ck.IsVersionKey = true
		if len(parts) < 5 {
			return fmt.Errorf("无效的版本时间戳 key 格式: %s", key)
		}
		ck.SystemPrefix = parts[1]
		companyUuid, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("无法解析公司 UUID: %v", err)
		}
		ck.CompanyUuid = companyUuid
		ck.ObjectType = parts[3]
		// 添加 object_uuid 和其他部分
		if len(parts) > 4 {
			ck.Parts = parts[4:]
		}
		return nil
	}

	// 普通缓存 key
	ck.IsVersionKey = false
	ck.SystemPrefix = parts[0]
	companyUuid, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("无法解析公司 UUID: %v", err)
	}
	ck.CompanyUuid = companyUuid
	ck.ObjectType = parts[2]
	// 添加其他部分
	if len(parts) > 3 {
		ck.Parts = parts[3:]
	}

	return nil
}

// String 重新构建缓存 key 字符串
// 返回：
//   - string: 构建后的缓存 key
func (ck *CacheKey) String() string {
	var parts []string

	if ck.IsVersionKey {
		// 版本时间戳 key: {cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
		parts = []string{
			CacheVersionPrefix,
			ck.SystemPrefix,
			fmt.Sprintf("%d", ck.CompanyUuid),
			ck.ObjectType,
		}
	} else {
		// 普通缓存 key: {system_prefix}:{company_uuid}:{object_type}:...
		parts = []string{
			ck.SystemPrefix,
			fmt.Sprintf("%d", ck.CompanyUuid),
			ck.ObjectType,
		}
	}

	// 添加其他部分
	if len(ck.Parts) > 0 {
		parts = append(parts, ck.Parts...)
	}

	return strings.Join(parts, ":")
}

// GetObjectUuid 从缓存 key 中提取对象 UUID
// 注意：
//   - 对于 product_list 对象，可变部分有多个（source, pageNo, pageSize 等），不支持提取 object_uuid
//   - 对于其他对象，可变部分就是 object_uuid（单个部分）
//
// 参数：
//   - 无
//
// 返回：
//   - uint64: 对象 UUID（product_list 返回 0）
//   - error: 错误信息（product_list 返回错误，表示不支持）
func (ck *CacheKey) GetObjectUuid() (uint64, error) {
	// product_list 对象不支持提取 object_uuid
	if ck.ObjectType == ObjectTypeProductList {
		return 0, fmt.Errorf("product_list 对象不支持提取 object_uuid，可变部分有多个")
	}

	// 检查是否有可变部分
	if len(ck.Parts) == 0 {
		return 0, fmt.Errorf("缓存 key 中没有可变部分，无法提取 object_uuid")
	}

	// 解析第一个可变部分为 object_uuid
	objectUuid, err := strconv.ParseUint(ck.Parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析 object_uuid: %v", err)
	}

	return objectUuid, nil
}

func NewCacheKey(key string) (*CacheKey, error) {
	ck := &CacheKey{}
	if err := ck.Parse(key); err != nil {
		return nil, fmt.Errorf("解析缓存 key 失败: %v", err)
	}
	return ck, nil
}
