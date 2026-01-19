package model

import "time"

// CacheHitRateSnapshot 缓存命中率快照
type CacheHitRateSnapshot struct {
	ID           uint64    `gorm:"primaryKey;column:id;type:bigint unsigned;not null;autoIncrement" json:"id"`
	InstanceID   string    `gorm:"column:instance_id;type:varchar(128);not null;default:'';index:idx_instance_snapshot" json:"instance_id"`
	SnapshotTime time.Time `gorm:"column:snapshot_time;type:datetime;not null;index:idx_snapshot_time" json:"snapshot_time"`
	Hits         int64     `gorm:"column:hits;type:bigint unsigned;not null;default:0" json:"hits"`
	Misses       int64     `gorm:"column:misses;type:bigint unsigned;not null;default:0" json:"misses"`
	Total        int64     `gorm:"column:total;type:bigint unsigned;not null;default:0" json:"total"`
	HitRate      float64   `gorm:"column:hit_rate;type:decimal(5,2);not null;default:0.00" json:"hit_rate"`
	KeyCount     int       `gorm:"column:key_count;type:int unsigned;not null;default:0" json:"key_count"`
	KeyStats     string    `gorm:"column:key_stats;type:json;default:null" json:"key_stats"`
	IsRestart    int       `gorm:"column:is_restart;type:tinyint(1) unsigned;not null;default:0;index:idx_instance_restart" json:"is_restart"`
	CreateTime   int64     `gorm:"column:create_time;type:bigint unsigned;not null;default:0" json:"create_time"`
}

// TableName 指定表名
func (CacheHitRateSnapshot) TableName() string {
	return "ttpos_cache_hit_rate_snapshot"
}
