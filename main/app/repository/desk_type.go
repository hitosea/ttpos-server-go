package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDeskTypeRepo 桌台
type IDeskTypeRepo interface {
	GetDeskTypeList() ([]model.DeskType, error)
	UpdateDeskType(uuid uint, deskType model.DeskType) error
	CreateDeskType(deskType model.DeskType) (uint64, error)
	DeleteDeskType(uuid uint) error
}

func NewDeskTypeRepo(db *gorm.DB) IDeskTypeRepo {
	return NewDeskTypeRepoImpl(db)
}

// NewDeskTypeRepoImpl 创建新的商品规格仓库实现
func NewDeskTypeRepoImpl(db *gorm.DB) *DeskTypeRepoImpl {
	return &DeskTypeRepoImpl{db: db}
}

type DeskTypeRepoImpl struct {
	db *gorm.DB
}

// GetDeskTypeList 获取桌台类型列表，排除逻辑删除的桌台类型
func (r *DeskTypeRepoImpl) GetDeskTypeList() ([]model.DeskType, error) {
	var deskTypes []model.DeskType
	err := r.db.Model(&model.DeskType{}).Where("delete_time = ?", 0).Find(&deskTypes).Error
	return deskTypes, err
}

// UpdateDeskType 更新桌台类型
func (r *DeskTypeRepoImpl) UpdateDeskType(uuid uint, deskType model.DeskType) error {
	if err := r.db.Model(&model.DeskType{}).Where("uuid = ?", uuid).Updates(deskType).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateDeskType 创建桌台类型
func (r *DeskTypeRepoImpl) CreateDeskType(deskType model.DeskType) (uint64, error) {
	// 创建桌台类型
	if err := r.db.Model(&model.DeskType{}).Create(&deskType).Error; err != nil {
		return 0, err
	}
	return deskType.Uuid, nil
}

// DeleteDeskType 软删除商品规格
func (r *DeskTypeRepoImpl) DeleteDeskType(id uint) error {
	return r.db.Model(&model.DeskType{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
