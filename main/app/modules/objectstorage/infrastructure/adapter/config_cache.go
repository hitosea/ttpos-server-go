package adapter

import (
	"encoding/json"
	"slices"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// SettingObjectStorageCache 对象存储缓存配置的 key（存储在 saas 库的 ttpos_setting 表中）
const SettingObjectStorageCache = "object_storage_cache"

// ObjectStorageCacheConfig 对象存储缓存配置结构
type ObjectStorageCacheConfig struct {
	Enabled   bool     `json:"enabled"`   // 是否启用缓存（全局开关）
	Whitelist []uint64 `json:"whitelist"` // 白名单（门店 UUID 列表）
}

// 默认配置值（用于向后兼容）
var defaultObjectStorageCacheConfig = ObjectStorageCacheConfig{
	Enabled: false,
	Whitelist: []uint64{
		7709131161600000,
	},
}

// 配置缓存层（使用三级缓存基础设施）
var configCacheLayer repository.CacheLayer[ObjectStorageCacheConfig]

// InitObjectStorageCacheConfigCache 初始化对象存储缓存配置的缓存层
// 参数：
//   - underlyingCache: 底层缓存实例
//   - dbGetter: 数据库获取函数（用于查询 saas 库的 ttpos_setting 表）
func InitObjectStorageCacheConfigCache(underlyingCache cache.Cache, dbGetter func() *gorm.DB) {
	// 创建缓存组配置：开启 L1 缓存，关闭 L2 缓存，L1 TTL 5 分钟
	groupConfig := cache.GroupConfig{
		Name:             "object-storage-cache-config",
		EnableLocalCache: true,             // 开启 L1 本地缓存
		EnableRedisCache: false,            // 关闭 L2 Redis 缓存
		NegativeTTL:      30 * time.Second, // 开启负缓存
		L1TTL:            5 * time.Minute,  // L1 缓存 5 分钟
		L2TTL:            0,                // L2 缓存不启用
	}

	// 创建缓存层（使用单例模式）
	configCacheLayer = GetOrCreateCacheLayer[ObjectStorageCacheConfig](
		groupConfig,
		underlyingCache,
		5*time.Minute, // 默认 TTL 5 分钟
		WithSingletonKey("object_storage_cache_config"),
		WithL1TTL(5*time.Minute), // L1 缓存 5 分钟
	)

	// 设置数据库查询函数（用于缓存未命中时查询数据库）
	setConfigDBGetter(dbGetter)
}

// dbGetterFunc 数据库获取函数（用于查询配置）
var dbGetterFunc func() *gorm.DB

// setConfigDBGetter 设置数据库获取函数
func setConfigDBGetter(getter func() *gorm.DB) {
	dbGetterFunc = getter
}

// GetObjectStorageCacheConfig 获取对象存储缓存配置（使用三级缓存基础设施）
func GetObjectStorageCacheConfig() ObjectStorageCacheConfig {
	if configCacheLayer == nil {
		// 缓存层未初始化，返回默认值
		return defaultObjectStorageCacheConfig
	}

	// 使用缓存层查询配置
	config, err := configCacheLayer.GET(
		SettingObjectStorageCache,
		func() (ObjectStorageCacheConfig, error) {
			// 缓存未命中时的查询函数
			return queryConfigFromDB()
		},
	)

	if err != nil {
		// 查询失败，返回默认值
		return defaultObjectStorageCacheConfig
	}

	return config
}

// queryConfigFromDB 从数据库查询配置
func queryConfigFromDB() (ObjectStorageCacheConfig, error) {
	if dbGetterFunc == nil {
		return defaultObjectStorageCacheConfig, nil
	}

	db := dbGetterFunc()
	if db == nil {
		return defaultObjectStorageCacheConfig, nil
	}

	// 从数据库读取配置
	var setting struct {
		Key    string
		Values string
	}
	err := db.Table("ttpos_setting").Where("`key` = ? AND delete_time = 0", SettingObjectStorageCache).First(&setting).Error
	if err != nil {
		// 查询失败或记录不存在，返回默认值
		return defaultObjectStorageCacheConfig, nil
	}

	if setting.Key == "" {
		// 配置不存在，返回默认值
		return defaultObjectStorageCacheConfig, nil
	}

	// 解析 JSON 配置
	var config ObjectStorageCacheConfig
	if err := json.Unmarshal([]byte(setting.Values), &config); err != nil {
		// 解析失败，返回默认值
		return defaultObjectStorageCacheConfig, nil
	}

	return config, nil
}

// IsObjectStorageCacheEnabled 检查是否启用对象存储缓存
// 需要同时满足：
//  1. enabled 为 true（全局开关）
//  2. 当前门店 UUID 在白名单内（如果白名单不为空）
//
// 参数：
//   - companyUuid: 门店 UUID
//
// 返回：
//   - true: 启用缓存，false: 禁用缓存
func IsObjectStorageCacheEnabled(companyUuid uint64) bool {
	config := GetObjectStorageCacheConfig()

	// 如果全局开关关闭，直接返回 false
	if !config.Enabled {
		return false
	}

	// 如果白名单为空，则不启用缓存
	if len(config.Whitelist) == 0 {
		return false
	}

	// 检查当前门店是否在白名单内
	return slices.Contains(config.Whitelist, companyUuid)
}
