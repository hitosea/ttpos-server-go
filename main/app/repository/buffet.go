package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetRepo 桌台
type IBuffetRepo interface {
	GetBuffetList(pageNo, pageSize int, opts ...DBOption) ([]*model.BuffetPackage, int64, error)
	GetBuffetListNoPage(opts ...DBOption) ([]*model.BuffetPackage, error)
	GetBuffetListInDeskOpen(pageNo, pageSize int) ([]*model.BuffetPackage, int64, error) // 获取自助餐列表（桌台开台）
	GetBuffetListByUuids(uuids []uint64) ([]*model.BuffetPackage, error)                 // 获取自助餐列表通过UUID列表，用于开台写入顾客数据
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

func (r *BuffetRepoImpl) GetBuffetListNoPage(opts ...DBOption) ([]*model.BuffetPackage, error) {
	list, _, err := r.GetBuffetList(1, 1000000, opts...)
	return list, err
}

// GetBuffetListInDeskOpen 获取自助餐列表（桌台开台）
func (r *BuffetRepoImpl) GetBuffetListInDeskOpen(pageNo, pageSize int) ([]*model.BuffetPackage, int64, error) {
	return r.GetBuffetList(pageNo, pageSize,
		CommonRepo.Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
		CommonRepo.Preload(WithPreload{
			Query: "BuffetCustomerTypePrices",
			Args: []interface{}{
				CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
			},
		}),
		CommonRepo.Preload(WithPreload{
			Query: "BuffetCustomerTypePrices.BuffetCustomerType",
		}),
	)
}

func (r *BuffetRepoImpl) GetBuffetListByUuids(uuids []uint64) ([]*model.BuffetPackage, error) {
	return r.GetBuffetListNoPage(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "Tax",
			},
		),
		CommonRepo.Preload(
			WithPreload{
				Query: "BuffetProducts",
			},
			WithPreload{
				Query: "Tax",
			},
		),
		CommonRepo.Preload(WithPreload{
			Query: "BuffetCustomerTypePrices",
			Args: []interface{}{
				CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
			},
		}),
		CommonRepo.Preload(WithPreload{
			Query: "BuffetCustomerTypePrices.BuffetCustomerType",
		}))
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
