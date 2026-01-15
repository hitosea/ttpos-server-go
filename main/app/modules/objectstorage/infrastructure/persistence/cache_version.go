package persistence

import (
	goCtx "context"
	"fmt"
	"strconv"
	"time"

	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheVersionManager 缓存版本时间戳管理器
// 用于管理缓存对象的版本时间戳，实现缓存失效机制
type CacheVersionManager struct {
	cacheInstance cache.Cache
	companyWide   bool
}

// NewCacheVersionManager 创建缓存版本时间戳管理器
// 参数：
//   - cacheInstance: 缓存实例（用于获取 Redis 客户端）
//
// 返回：
//   - *CacheVersionManager: 缓存版本时间戳管理器实例
func NewCacheVersionManager(cacheInstance cache.Cache) *CacheVersionManager {
	return &CacheVersionManager{
		cacheInstance: cacheInstance,
	}
}

// GetCacheVersionTimestampOption GetCacheVersionTimestamp 方法的选项
type GetCacheVersionTimestampOption struct {
	// CompanyWide 是否在公司范围内使对象失效
	// 当为 true 时，无论 objectUuid 为什么值都赋值为 0（表示全局版本时间戳）
	CompanyWide bool
}

// WithCompanyWide 设置是否在公司范围内使对象失效（用于 GetCacheVersionTimestamp）
// 参数：
//   - companyWide: 如果为 true，则无论 objectUuid 为什么值都赋值为 0
func WithCompanyWide(companyWide bool) func(*GetCacheVersionTimestampOption) {
	return func(opt *GetCacheVersionTimestampOption) {
		opt.CompanyWide = companyWide
	}
}

func (cvm *CacheVersionManager) SetCompanyWide(companyWide bool) {
	cvm.companyWide = companyWide
}

// GetCacheVersionTimestamp 获取缓存版本时间戳
// 参数：
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型
//   - objectUuid: 对象 UUID（通常为 0，表示全局版本时间戳）
//   - opts: 选项函数（可选），如 WithCompanyWide(true) 表示在公司范围内使对象失效
//
// 返回：
//   - int64: 版本时间戳（Unix 时间戳，秒）
//   - bool: 是否存在版本时间戳
func (cvm *CacheVersionManager) GetCacheVersionTimestamp(companyUuid uint64, objectType string, objectUuid uint64, opts ...func(*GetCacheVersionTimestampOption)) (int64, bool) {
	// 解析选项
	option := &GetCacheVersionTimestampOption{
		CompanyWide: false, // 默认值
	}
	for _, opt := range opts {
		opt(option)
	}

	// 如果 CompanyWide 为 true，无论 objectUuid 为什么值都赋值为 0
	if option.CompanyWide {
		objectUuid = 0
	}

	// 如果 companyWide 为 true，无论 objectUuid 为什么值都赋值为 0
	if cvm.companyWide {
		objectUuid = 0
	}

	// 获取 Redis 客户端
	var client redis.UniversalClient
	if clusterClient := cvm.cacheInstance.GetClusterClient(); clusterClient != nil {
		client = clusterClient
	} else if redisClient := cvm.cacheInstance.GetClient(); redisClient != nil {
		client = redisClient
	} else {
		// 如果不是 Redis 缓存，返回 false
		return 0, false
	}

	// 构建版本时间戳 key（objectUuid 为 0 表示全局版本时间戳）
	versionKey := BuildCacheVersionKey(companyUuid, objectType, objectUuid)

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
//   - companyUuid: 门店 UUID
//   - objectType: 对象类型
//   - objectUuid: 对象 UUID（通常为 0，表示全局版本时间戳）
//
// 返回：
//   - error: 错误信息
func (cvm *CacheVersionManager) UpdateCacheVersionTimestamp(companyUuid uint64, objectType string, objectUuid uint64) error {
	if cvm.companyWide {
		objectUuid = 0
	}
	// 获取 Redis 客户端
	var client redis.UniversalClient
	if clusterClient := cvm.cacheInstance.GetClusterClient(); clusterClient != nil {
		client = clusterClient
	} else if redisClient := cvm.cacheInstance.GetClient(); redisClient != nil {
		client = redisClient
	} else {
		// 如果不是 Redis 缓存，无法设置版本时间戳，直接返回
		return nil
	}

	// 构建版本时间戳 key（objectUuid 为 0 表示全局版本时间戳）
	versionKey := BuildCacheVersionKey(companyUuid, objectType, objectUuid)

	// 使用当前时间戳（秒）作为版本时间戳
	currentTimestamp := time.Now().Unix()

	// 设置版本时间戳，过期时间应该和最长有效的缓存时间一致（如 L2TTL）
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	backgroundCtx := goCtx.Background()
	if err := client.Set(backgroundCtx, versionKey, currentTimestamp, CacheVersionTTL).Err(); err != nil {
		return fmt.Errorf("设置缓存版本时间戳失败: %w", err)
	}

	logger.Logger.Debug("CacheVersionTimestamp update", zap.Int64("currentTimestamp", currentTimestamp), zap.String("versionKey", versionKey), zap.Duration("ttl", CacheVersionTTL))
	return nil
}
