package constant

import "slices"

// 缓存相关常量

// ObjectStorageCacheEnabled 是否启用对象存储缓存（全局开关）
// true: 启用缓存，false: 禁用缓存（直接查询数据库）
// 默认值：true（启用缓存）
const ObjectStorageCacheEnabled = true

// ObjectStorageCacheWhitelist 对象存储缓存白名单（门店 UUID 列表）
// 只有当 ObjectStorageCacheEnabled 为 true 且当前门店 UUID 在此白名单内时，才会启用缓存
// 如果白名单为空，则所有门店都启用缓存（当 ObjectStorageCacheEnabled 为 true 时）
var ObjectStorageCacheWhitelist = []uint64{
	// 示例：7709131161600000,
	// 可以在这里添加允许使用缓存的门店 UUID
	7709131161600000,
}

// IsObjectStorageCacheEnabled 检查是否启用对象存储缓存
// 需要同时满足：
//  1. ObjectStorageCacheEnabled 为 true（全局开关）
//  2. 当前门店 UUID 在白名单内（如果白名单不为空）
//
// 参数：
//   - companyUuid: 门店 UUID
//
// 返回：
//   - true: 启用缓存，false: 禁用缓存
func IsObjectStorageCacheEnabled(companyUuid uint64) bool {
	// 如果全局开关关闭，直接返回 false
	if !ObjectStorageCacheEnabled {
		return false
	}

	// 如果白名单为空，则不启用缓存
	if len(ObjectStorageCacheWhitelist) == 0 {
		return false
	}

	// 检查当前门店是否在白名单内
	return slices.Contains(ObjectStorageCacheWhitelist, companyUuid)
}
