package repository

import (
	"ttpos-server-go/app/model"
	appRepo "ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// ICommonRepo 通用
type ICommonRepo interface {
}

func NewCommonRepo(db *gorm.DB) ICommonRepo {
	return NewCommonRepoImpl(db)
}

// NewCommonRepoImpl 创建新的角色仓库实现
func NewCommonRepoImpl(db *gorm.DB) *CommonRepoImpl {
	return &CommonRepoImpl{db: db}
}

type CommonRepoImpl struct {
	db *gorm.DB
}

// 创建商品包数据
func (s *CommonRepoImpl) CreateProductPackage(productPackage *model.ProductPackage) error {
	productPackageRepo := appRepo.NewProductPackageRepo(s.db)
	return productPackageRepo.CreateProductPackage(productPackage)
}
