package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetRepo 桌台
type IBuffetRepo interface {
	GetBuffetList(pageNo, pageSize int, opts ...DBOption) ([]model.BuffetPackage, int64, error)
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
func (r *BuffetRepoImpl) GetBuffetList(pageNo, pageSize int, opts ...DBOption) ([]model.BuffetPackage, int64, error) {
	var buffets []model.BuffetPackage
	var total int64

	db := r.db.Model(&model.BuffetPackage{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&buffets).Error
	return buffets, total, err
}
