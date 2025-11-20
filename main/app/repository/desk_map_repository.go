package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDeskMapLayoutRepo 桌台地图布局仓库接口
//
// 任务: story-admin-desktop-table-map Phase 2.4
// 需求: R1.1-R1.6
//
// @version v2.10.0
type IDeskMapLayoutRepo interface {
	IDeskMapLayoutQueryRepo
	CreateLayout(layout model.DeskMapLayout) (uint64, error)
	UpdateLayout(areaUuid uint64, layout model.DeskMapLayout) error
	DeleteLayout(areaUuid uint64) error
}

// IDeskMapLayoutQueryRepo 桌台地图布局查询接口
type IDeskMapLayoutQueryRepo interface {
	FindByAreaUuid(areaUuid uint64) (*model.DeskMapLayout, error)
	FindAll() ([]model.DeskMapLayout, error)
}

// NewDeskMapLayoutRepo 创建桌台地图布局仓库
func NewDeskMapLayoutRepo(db *gorm.DB) IDeskMapLayoutRepo {
	return NewDeskMapLayoutRepoImpl(db)
}

// NewDeskMapLayoutRepoImpl 创建桌台地图布局仓库实现
func NewDeskMapLayoutRepoImpl(db *gorm.DB) *DeskMapLayoutRepoImpl {
	return &DeskMapLayoutRepoImpl{db: db}
}

// DeskMapLayoutRepoImpl 桌台地图布局仓库实现
type DeskMapLayoutRepoImpl struct {
	db *gorm.DB
}

// FindByAreaUuid 根据区域UUID查找布局
func (r *DeskMapLayoutRepoImpl) FindByAreaUuid(areaUuid uint64) (*model.DeskMapLayout, error) {
	var layout model.DeskMapLayout
	err := r.db.Model(&model.DeskMapLayout{}).
		Where("area_uuid = ? AND delete_time = ?", areaUuid, 0).
		First(&layout).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录返回 nil，不报错
		}
		return nil, errors.WithMessage(err)
	}
	
	return &layout, nil
}

// FindAll 获取所有布局（未删除的）
func (r *DeskMapLayoutRepoImpl) FindAll() ([]model.DeskMapLayout, error) {
	var layouts []model.DeskMapLayout
	err := r.db.Model(&model.DeskMapLayout{}).
		Where("delete_time = ?", 0).
		Find(&layouts).Error
	
	return layouts, errors.WithMessage(err)
}

// CreateLayout 创建布局
func (r *DeskMapLayoutRepoImpl) CreateLayout(layout model.DeskMapLayout) (uint64, error) {
	if err := r.db.Model(&model.DeskMapLayout{}).Create(&layout).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return layout.Uuid, nil
}

// UpdateLayout 更新布局
func (r *DeskMapLayoutRepoImpl) UpdateLayout(areaUuid uint64, layout model.DeskMapLayout) error {
	err := r.db.Model(&model.DeskMapLayout{}).
		Where("area_uuid = ? AND delete_time = ?", areaUuid, 0).
		Updates(map[string]interface{}{
			"layout_json": layout.LayoutJson,
			"update_time": layout.UpdateTime,
		}).Error
	
	return errors.WithMessage(err)
}

// DeleteLayout 软删除布局
func (r *DeskMapLayoutRepoImpl) DeleteLayout(areaUuid uint64) error {
	err := r.db.Model(&model.DeskMapLayout{}).
		Where("area_uuid = ? AND delete_time = ?", areaUuid, 0).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
	
	return errors.WithMessage(err)
}

