package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDeskRegionRepo 桌台区域
type IDeskRegionRepo interface {
	GetDeskRegionList() ([]model.DeskRegion, error)
	UpdateDeskRegion(uuid uint, deskRegion model.DeskRegion) error
	CreateDeskRegion(deskRegion model.DeskRegion) (uint, error)
	DeleteDeskRegion(uuid uint) error
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
func (r *DeskRegionRepoImpl) GetDeskRegionList() ([]model.DeskRegion, error) {
	var deskRegions []model.DeskRegion
	err := r.db.Model(&model.DeskRegion{}).Where("delete_time = ?", 0).Find(&deskRegions).Error
	return deskRegions, err
}

// UpdateDeskRegion 更新桌台区域
func (r *DeskRegionRepoImpl) UpdateDeskRegion(uuid uint, deskRegion model.DeskRegion) error {
	if err := r.db.Model(&model.DeskRegion{}).Where("uuid = ?", uuid).Updates(deskRegion).Error; err != nil {
		return err
	}
	return nil
}

// CreateDeskRegion 创建桌台区域
func (r *DeskRegionRepoImpl) CreateDeskRegion(deskRegion model.DeskRegion) (uint, error) {
	// 创建桌台区域
	if err := r.db.Create(&deskRegion).Error; err != nil {
		return 0, err
	}
	return deskRegion.Uuid, nil
}

// DeleteDeskRegion 软删除桌台区域
func (r *DeskRegionRepoImpl) DeleteDeskRegion(id uint) error {
	return r.db.Model(&model.DeskRegion{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
