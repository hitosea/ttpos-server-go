package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetRepo 桌台
type IBuffetRepo interface {
	GetBuffetList(pageNo, pageSize int) ([]model.BuffetPackage, int64, error)
}

func NewBuffetRepo(db *gorm.DB) IBuffetRepo {
	return NewBuffetRepoImpl(db)
}

// NewBuffetRepoImpl 创建新的桌台仓库实现
func NewBuffetRepoImpl(db *gorm.DB) *BuffetRepoImpl {
	return &BuffetRepoImpl{db: db}
}

type BuffetRepoImpl struct {
	db *gorm.DB
}

// GetBuffetList 获取
func (r *BuffetRepoImpl) GetBuffetList(pageNo, pageSize int) ([]model.BuffetPackage, int64, error) {
	var buffets []model.BuffetPackage
	var total int64

	query := r.db.Model(&model.BuffetPackage{}).Preload("BuffetCustomerTypePrice.BuffetCustomerType.MultiLanguageName").Preload("MultiLanguageName").Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&buffets).Error
	return buffets, total, err
}
