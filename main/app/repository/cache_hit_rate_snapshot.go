package repository

import (
	"errors"
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ICacheHitRateSnapshotRepo 缓存命中率快照仓库接口
type ICacheHitRateSnapshotRepo interface {
	// SaveSnapshot 保存快照
	SaveSnapshot(snapshot *model.CacheHitRateSnapshot) error
	// GetSnapshotsByTimeRange 按时间范围查询快照
	GetSnapshotsByTimeRange(instanceID string, startTime, endTime time.Time) ([]*model.CacheHitRateSnapshot, error)
	// GetLatestSnapshot 获取指定实例的最新快照
	GetLatestSnapshot(instanceID string) (*model.CacheHitRateSnapshot, error)
}

// CacheHitRateSnapshotRepoImpl 缓存命中率快照仓库实现
type CacheHitRateSnapshotRepoImpl struct {
	db *gorm.DB
}

// NewCacheHitRateSnapshotRepo 创建缓存命中率快照仓库
func NewCacheHitRateSnapshotRepo(db *gorm.DB) ICacheHitRateSnapshotRepo {
	return &CacheHitRateSnapshotRepoImpl{
		db: db,
	}
}

// SaveSnapshot 保存快照
func (r *CacheHitRateSnapshotRepoImpl) SaveSnapshot(snapshot *model.CacheHitRateSnapshot) error {
	return r.db.Create(snapshot).Error
}

// GetSnapshotsByTimeRange 按时间范围查询快照
func (r *CacheHitRateSnapshotRepoImpl) GetSnapshotsByTimeRange(instanceID string, startTime, endTime time.Time) ([]*model.CacheHitRateSnapshot, error) {
	var snapshots []*model.CacheHitRateSnapshot
	query := r.db.Model(&model.CacheHitRateSnapshot{}).
		Where("snapshot_time >= ? AND snapshot_time <= ?", startTime, endTime)

	if instanceID != "" {
		query = query.Where("instance_id = ?", instanceID)
	}

	err := query.Order("snapshot_time ASC").Find(&snapshots).Error
	return snapshots, err
}

// GetLatestSnapshot 获取指定实例的最新快照
func (r *CacheHitRateSnapshotRepoImpl) GetLatestSnapshot(instanceID string) (*model.CacheHitRateSnapshot, error) {
	var snapshot model.CacheHitRateSnapshot
	err := r.db.Model(&model.CacheHitRateSnapshot{}).
		Where("instance_id = ?", instanceID).
		Order("snapshot_time DESC").
		First(&snapshot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &snapshot, nil
}
