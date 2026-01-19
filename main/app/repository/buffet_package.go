package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetPackageRepo 自助餐套餐
type IBuffetPackageRepo interface {
	GetBuffetPackage(opts ...DBOption) (model.BuffetPackage, error)                                       // 获取自助餐套餐
	GetBuffetPackageByUuidWithAssociations(buffetPackageUuid uint64) (*model.BuffetPackage, error)        // 通过UUID查询自助餐信息，包括MultiLanguageName、BuffetCustomerTypePrices、BuffetProducts.ProductPackage
	GetBuffetPackagesByUuidsWithAssociations(buffetPackageUuids []uint64) ([]*model.BuffetPackage, error) // 批量通过UUID列表查询自助餐信息，包括MultiLanguageName、BuffetCustomerTypePrices、BuffetProducts.ProductPackage
}

func NewBuffetPackageRepo(db *gorm.DB) IBuffetPackageRepo {
	return NewBuffetPackageRepoImpl(db)
}

// NewBuffetPackageRepoImpl 创建新的自助餐套餐仓库实现
func NewBuffetPackageRepoImpl(db *gorm.DB) *BuffetPackageRepoImpl {
	return &BuffetPackageRepoImpl{db: db}
}

type BuffetPackageRepoImpl struct {
	db *gorm.DB
}

// WhereUuid 通过uuid查询
func (r *BuffetPackageRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// GetBuffetPackage 获取自助餐套餐
func (r *BuffetPackageRepoImpl) GetBuffetPackage(opts ...DBOption) (model.BuffetPackage, error) {
	var buffetPackage model.BuffetPackage
	db := r.db.Model(&model.BuffetPackage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&buffetPackage).Error
	return buffetPackage, errors.WithMessage(err)
}

// GetBuffetPackageByUuidWithAssociations 通过UUID查询自助餐信息，包括MultiLanguageName、BuffetCustomerTypePrices、BuffetProducts.ProductPackage
func (r *BuffetPackageRepoImpl) GetBuffetPackageByUuidWithAssociations(buffetPackageUuid uint64) (*model.BuffetPackage, error) {
	var buffetPackage model.BuffetPackage
	err := r.db.Model(&model.BuffetPackage{}).
		Where("uuid = ?", buffetPackageUuid).
		Preload("MultiLanguageName").
		Preload("BuffetCustomerTypePrices").
		Preload("BuffetProducts.ProductPackage").
		First(&buffetPackage).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &buffetPackage, nil
}

// GetBuffetPackagesByUuidsWithAssociations 批量通过UUID列表查询自助餐信息，包括MultiLanguageName、BuffetCustomerTypePrices、BuffetProducts.ProductPackage
func (r *BuffetPackageRepoImpl) GetBuffetPackagesByUuidsWithAssociations(buffetPackageUuids []uint64) ([]*model.BuffetPackage, error) {
	if len(buffetPackageUuids) == 0 {
		return []*model.BuffetPackage{}, nil
	}
	var buffetPackages []*model.BuffetPackage
	err := r.db.Model(&model.BuffetPackage{}).
		Where("uuid IN (?)", buffetPackageUuids).
		Preload("MultiLanguageName").
		Preload("BuffetCustomerTypePrices").
		Preload("BuffetProducts.ProductPackage").
		Find(&buffetPackages).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return buffetPackages, nil
}
