package resp

import "time"

// CacheHitRateKeyStatsResp 单个 key 的统计信息响应
type CacheHitRateKeyStatsResp struct {
	Key        string    `json:"key"`         // key名称
	Hits       int64     `json:"hits"`        // 命中次数
	Misses     int64     `json:"misses"`      // 未命中次数
	HitRate    float64   `json:"hit_rate"`    // 命中率
	Total      int64     `json:"total"`       // 总请求数
	LastAccess time.Time `json:"last_access"` // 最后访问时间
}

// CacheHitRateStatsDataResp 缓存统计信息响应
type CacheHitRateStatsDataResp struct {
	Hits       int64     `json:"hits"`        // 命中次数
	Misses     int64     `json:"misses"`      // 未命中次数
	HitRate    float64   `json:"hit_rate"`    // 命中率
	Total      int64     `json:"total"`       // 总请求数
	LastUpdate time.Time `json:"last_update"` // 最后更新时间
	KeyCount   int       `json:"key_count"`   // Key数量
}

// CacheHitRateStatsResp 缓存命中率统计响应
type CacheHitRateStatsResp struct {
	Stats   CacheHitRateStatsDataResp  `json:"stats"`    // 总体统计信息
	TopKeys []CacheHitRateKeyStatsResp `json:"top_keys"` // Top N 的 key 统计信息
}
