package persistence

import (
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

// ITakeoutOrderMaterialRepo 外卖订单原料仓储接口
type ITakeoutOrderMaterialRepo interface {
	// Create 创建外卖订单原料记录
	Create(takeoutOrderMaterial *takeoutModel.TakeoutOrderMaterial) error
	// BatchCreate 批量创建外卖订单原料记录
	BatchCreate(takeoutOrderMaterials []*takeoutModel.TakeoutOrderMaterial) error
	// GetTakeoutOrderMaterialByCreateTimeBetween 根据创建时间区间查询外卖订单原料
	GetTakeoutOrderMaterialByCreateTimeBetween(startTime, endTime int64) ([]*takeoutModel.TakeoutOrderMaterial, error)
	// UpdateTakeoutOrderMaterialIsSummarized 批量更新外卖订单原料的统计状态
	UpdateTakeoutOrderMaterialIsSummarized(uuids []uint64) error
	// MarkTakeoutOrderMaterialsAsSummarized 标记外卖订单的所有原料为已汇总
	MarkTakeoutOrderMaterialsAsSummarized(takeoutOrderUuid uint64) error
	// GetByOrderUuid 根据订单UUID查询原料消耗记录（预加载原料信息）
	GetByOrderUuid(takeoutOrderUuid uint64) ([]*takeoutModel.TakeoutOrderMaterial, error)
	// DeleteByOrderUuid 根据订单UUID删除原料消耗记录（硬删除，用于订单更新时重新计算）
	DeleteByOrderUuid(takeoutOrderUuid uint64) error
}

type takeoutOrderMaterialRepoImpl struct {
	db *gorm.DB
}

// NewTakeoutOrderMaterialRepo 创建外卖订单原料仓储实例
func NewTakeoutOrderMaterialRepo(db *gorm.DB) ITakeoutOrderMaterialRepo {
	return &takeoutOrderMaterialRepoImpl{db: db}
}

// Create 创建外卖订单原料记录
func (r *takeoutOrderMaterialRepoImpl) Create(takeoutOrderMaterial *takeoutModel.TakeoutOrderMaterial) error {
	return r.db.Create(takeoutOrderMaterial).Error
}

// BatchCreate 批量创建外卖订单原料记录
func (r *takeoutOrderMaterialRepoImpl) BatchCreate(takeoutOrderMaterials []*takeoutModel.TakeoutOrderMaterial) error {
	if len(takeoutOrderMaterials) == 0 {
		return nil
	}
	return r.db.Create(takeoutOrderMaterials).Error
}

// GetTakeoutOrderMaterialByCreateTimeBetween 根据创建时间区间查询外卖订单原料
func (r *takeoutOrderMaterialRepoImpl) GetTakeoutOrderMaterialByCreateTimeBetween(startTime, endTime int64) ([]*takeoutModel.TakeoutOrderMaterial, error) {
	var takeoutOrderMaterials []*takeoutModel.TakeoutOrderMaterial
	err := r.db.Model(&takeoutModel.TakeoutOrderMaterial{}).
		Where("create_time BETWEEN ? AND ? AND delete_time = 0", startTime, endTime).
		Where("is_summarized = ?", 0). // 只查询未统计的记录
		Preload("Material", func(db *gorm.DB) *gorm.DB {
			return db.Select("uuid, name, unit_uuid, supplier_uuid").
				Preload("Unit", func(db *gorm.DB) *gorm.DB {
					return db.Select("uuid, name")
				})
		}).
		Find(&takeoutOrderMaterials).Error
	return takeoutOrderMaterials, err
}

// UpdateTakeoutOrderMaterialIsSummarized 批量更新外卖订单原料的统计状态
func (r *takeoutOrderMaterialRepoImpl) UpdateTakeoutOrderMaterialIsSummarized(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Model(&takeoutModel.TakeoutOrderMaterial{}).
		Where("uuid IN ?", uuids).
		Update("is_summarized", 1).Error
}

// MarkTakeoutOrderMaterialsAsSummarized 标记外卖订单的所有原料为已汇总
// 在取消订单时调用，避免被日终统计重复处理
func (r *takeoutOrderMaterialRepoImpl) MarkTakeoutOrderMaterialsAsSummarized(takeoutOrderUuid uint64) error {
	return r.db.Model(&takeoutModel.TakeoutOrderMaterial{}).
		Where("takeout_order_uuid = ?", takeoutOrderUuid).
		Where("is_summarized = ?", 0).
		Update("is_summarized", 1).Error
}

// GetByOrderUuid 根据订单UUID查询原料消耗记录（预加载原料信息）
// 用于构建出库清单时直接读取已保存的原料消耗记录
func (r *takeoutOrderMaterialRepoImpl) GetByOrderUuid(takeoutOrderUuid uint64) ([]*takeoutModel.TakeoutOrderMaterial, error) {
	var materials []*takeoutModel.TakeoutOrderMaterial
	err := r.db.Model(&takeoutModel.TakeoutOrderMaterial{}).
		Where("takeout_order_uuid = ?", takeoutOrderUuid).
		Where("delete_time = 0").
		Preload("Material").
		Find(&materials).Error
	return materials, err
}

// DeleteByOrderUuid 根据订单UUID删除原料消耗记录（硬删除，用于订单更新时重新计算）
// 注意：此方法仅用于订单更新场景，在事务中调用，删除旧记录后立即重新计算
func (r *takeoutOrderMaterialRepoImpl) DeleteByOrderUuid(takeoutOrderUuid uint64) error {
	return r.db.Unscoped().
		Where("takeout_order_uuid = ?", takeoutOrderUuid).
		Delete(&takeoutModel.TakeoutOrderMaterial{}).Error
}
