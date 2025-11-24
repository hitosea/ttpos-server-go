package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDeskRegionRepo 桌台区域
type IDeskRegionRepo interface {
	IDeskRegionQueryRepo
	UpdateDeskRegion(uuid uint, deskRegion model.DeskRegion) error
	CreateDeskRegion(deskRegion model.DeskRegion) (uint64, error)
	DeleteDeskRegion(uuid uint) error
}

// IDeskRegionQueryRepo 桌台区域查询
type IDeskRegionQueryRepo interface {
	GetDeskRegionList(opts ...DBOption) ([]model.DeskRegion, error)

	WithDeskMapLayout() DBOption
}

func NewDeskRegionRepo(db *gorm.DB) IDeskRegionRepo {
	return NewDeskRegionRepoImpl(db)
}

// NewDeskRegionRepoImpl 创建新的桌台区域仓库实现
func NewDeskRegionRepoImpl(db *gorm.DB) *DeskRegionRepoImpl {
	return &DeskRegionRepoImpl{db: db}
}

type DeskRegionRepoImpl struct {
	db *gorm.DB
}

// GetDeskRegionList 获取桌台区域列表，排除逻辑删除的桌台区域
func (r *DeskRegionRepoImpl) GetDeskRegionList(opts ...DBOption) ([]model.DeskRegion, error) {
	var deskRegions []model.DeskRegion
	db := r.db.Model(&model.DeskRegion{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("delete_time = ?", 0).Find(&deskRegions).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return deskRegions, nil
}

// UpdateDeskRegion 更新桌台区域
func (r *DeskRegionRepoImpl) UpdateDeskRegion(uuid uint, deskRegion model.DeskRegion) error {
	if err := r.db.Model(&model.DeskRegion{}).Where("uuid = ?", uuid).Updates(deskRegion).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateDeskRegion 创建桌台区域
func (r *DeskRegionRepoImpl) CreateDeskRegion(deskRegion model.DeskRegion) (uint64, error) {
	// 创建桌台区域
	if err := r.db.Model(&model.DeskRegion{}).Create(&deskRegion).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return deskRegion.Uuid, nil
}

// DeleteDeskRegion 软删除桌台区域
func (r *DeskRegionRepoImpl) DeleteDeskRegion(id uint) error {
	return r.db.Model(&model.DeskRegion{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

// WithDeskMapLayout 预加载桌台地图布局
func (r *DeskRegionRepoImpl) WithDeskMapLayout() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("DeskMapLayout", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0)
		})
	}
}
