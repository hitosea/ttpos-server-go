package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetRepo 桌台
type IBuffetRepo interface {
	GetBuffetList(pageNo, pageSize int, opts ...DBOption) ([]*model.BuffetPackage, int64, error)
	GetBuffetListInDeskOpen(pageNo, pageSize int) ([]*model.BuffetPackage, int64, error) // 获取自助餐列表（桌台开台）
	GetBuffetInfo(opts ...DBOption) (model.BuffetPackage, error)                         // 获取自助餐详情
	GetBuffetCustomerTypeInfo(opts ...DBOption) (model.BuffetCustomerType, error)        // 获取自助餐顾客类型详情
	GetBuffetCustomerTypePrice(opts ...DBOption) (model.BuffetCustomerTypePrice, error)  // 获取自助餐顾客类型价格
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
func (r *BuffetRepoImpl) GetBuffetList(pageNo, pageSize int, opts ...DBOption) ([]*model.BuffetPackage, int64, error) {
	var buffets []*model.BuffetPackage
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

// GetBuffetListInDeskOpen 获取自助餐列表（桌台开台）
func (r *BuffetRepoImpl) GetBuffetListInDeskOpen(pageNo, pageSize int) ([]*model.BuffetPackage, int64, error) {
	return r.GetBuffetList(pageNo, pageSize,
		NewCommonRepo().Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
		NewCommonRepo().Preload(WithPreload{
			Query: "BuffetCustomerTypePrices",
			Args: []interface{}{
				CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
			},
		}),
		NewCommonRepo().Preload(WithPreload{
			Query: "BuffetCustomerTypePrices.BuffetCustomerType",
		}),
	)
}

// GetBuffetInfo 获取自助餐详情
func (r *BuffetRepoImpl) GetBuffetInfo(opts ...DBOption) (model.BuffetPackage, error) {
	var buffet model.BuffetPackage

	db := r.db.Model(&model.BuffetPackage{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&buffet).Error

	return buffet, err
}

// GetBuffetCustomerTypeInfo 获取自助餐顾客类型详情
func (r *BuffetRepoImpl) GetBuffetCustomerTypeInfo(opts ...DBOption) (model.BuffetCustomerType, error) {
	var buffetCustomerType model.BuffetCustomerType

	db := r.db.Model(&model.BuffetCustomerType{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&buffetCustomerType).Error
	return buffetCustomerType, err
}

// GetBuffetCustomerTypePrice 获取自助餐顾客类型价格
func (r *BuffetRepoImpl) GetBuffetCustomerTypePrice(opts ...DBOption) (model.BuffetCustomerTypePrice, error) {
	var buffetCustomerTypePrice model.BuffetCustomerTypePrice

	db := r.db.Model(&model.BuffetCustomerTypePrice{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&buffetCustomerTypePrice).Error
	return buffetCustomerTypePrice, err
}
