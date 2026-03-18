package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IHqFieldOverrideRepo 总部字段覆盖追踪仓库接口
type IHqFieldOverrideRepo interface {
	// IsOverridden 检查指定实体的指定字段是否被子店覆盖
	IsOverridden(entityUuid uint64, fieldType string) bool

	// MarkOverridden 标记指定实体的指定字段已被子店覆盖（Upsert）
	MarkOverridden(entityUuid uint64, entityType, fieldType string) error

	// ClearOverride 清除指定实体的指定字段的覆盖标记
	ClearOverride(entityUuid uint64, fieldType string) error

	// ClearOverridesByEntityUuids 批量清除指定实体列表的指定字段的覆盖标记
	ClearOverridesByEntityUuids(entityUuids []uint64, fieldType string) error

	// BatchCheckOverridden 批量检查指定实体列表的指定字段是否被覆盖
	// 返回 map[entityUuid]bool
	BatchCheckOverridden(entityUuids []uint64, fieldType string) (map[uint64]bool, error)
}

// NewHqFieldOverrideRepo 创建仓库实例
func NewHqFieldOverrideRepo(db *gorm.DB) IHqFieldOverrideRepo {
	return &hqFieldOverrideRepo{db: db}
}

type hqFieldOverrideRepo struct {
	db *gorm.DB
}

func (r *hqFieldOverrideRepo) IsOverridden(entityUuid uint64, fieldType string) bool {
	var count int64
	r.db.Model(&model.HqFieldOverride{}).
		Where("entity_uuid = ? AND field_type = ? AND delete_time = 0", entityUuid, fieldType).
		Count(&count)
	return count > 0
}

func (r *hqFieldOverrideRepo) MarkOverridden(entityUuid uint64, entityType, fieldType string) error {
	var existing model.HqFieldOverride
	err := r.db.Model(&model.HqFieldOverride{}).
		Where("entity_uuid = ? AND field_type = ? AND delete_time = 0", entityUuid, fieldType).
		First(&existing).Error

	if err == nil && existing.ID > 0 {
		// 已存在，不需要重复创建
		return nil
	}

	// 创建新记录
	override := model.HqFieldOverride{
		EntityType: entityType,
		EntityUuid: entityUuid,
		FieldType:  fieldType,
	}
	return r.db.Create(&override).Error
}

func (r *hqFieldOverrideRepo) ClearOverride(entityUuid uint64, fieldType string) error {
	return r.db.Model(&model.HqFieldOverride{}).
		Where("entity_uuid = ? AND field_type = ? AND delete_time = 0", entityUuid, fieldType).
		Update("delete_time", time.Now().Unix()).Error
}

func (r *hqFieldOverrideRepo) ClearOverridesByEntityUuids(entityUuids []uint64, fieldType string) error {
	if len(entityUuids) == 0 {
		return nil
	}
	return r.db.Model(&model.HqFieldOverride{}).
		Where("entity_uuid IN ? AND field_type = ? AND delete_time = 0", entityUuids, fieldType).
		Update("delete_time", time.Now().Unix()).Error
}

func (r *hqFieldOverrideRepo) BatchCheckOverridden(entityUuids []uint64, fieldType string) (map[uint64]bool, error) {
	result := make(map[uint64]bool)
	if len(entityUuids) == 0 {
		return result, nil
	}

	var overrides []model.HqFieldOverride
	err := r.db.Model(&model.HqFieldOverride{}).
		Where("entity_uuid IN ? AND field_type = ? AND delete_time = 0", entityUuids, fieldType).
		Find(&overrides).Error
	if err != nil {
		return nil, err
	}

	for _, o := range overrides {
		result[o.EntityUuid] = true
	}
	return result, nil
}
