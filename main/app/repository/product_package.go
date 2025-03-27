package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageRepo interface {
	GetProductPackage(opts ...DBOption) (*model.ProductPackage, error)
	GetProductPackageBaseInfoByBomUuid(flavorBomUuid uint64) (*model.ProductBom, error)
	GetProductPackageListByUuids(uuids []uint64) ([]*model.ProductPackage, error)
	CreateProductPackage(productPackage *model.ProductPackage) error
}

type productPackageRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageRepo(db *gorm.DB) IProductPackageRepo {
	return &productPackageRepoImpl{db: db}
}

func (r *productPackageRepoImpl) GetProductPackage(opts ...DBOption) (*model.ProductPackage, error) {
	var productPackage model.ProductPackage
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackage)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackage, nil
}

func (r *productPackageRepoImpl) GetProductPackageBaseInfoByBomUuid(flavorBomUuid uint64) (*model.ProductBom, error) {
	// 先查询出ProductBom
	productBomRepo := NewProductBomRepo(r.db)
	productBom, err := productBomRepo.GetProductBom(
		CommonRepo.WhereByUuid(flavorBomUuid),
		CommonRepo.Preload(WithPreload{Query: "ProductPackage.TakeoutTax"}),
		CommonRepo.Preload(WithPreload{Query: "ProductPackage.DineTax"}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

func (r *productPackageRepoImpl) GetProductPackageListByUuids(uuids []uint64) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ?", uuids).Find(&productPackages).Error
	return productPackages, errors.WithMessage(err)
}

func (r *productPackageRepoImpl) CreateProductPackage(productPackage *model.ProductPackage) error {
	// 创建product_package表数据
	if err := r.db.Model(&model.ProductPackage{}).Create(productPackage).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 创建product_package_attribute_group表数据
	if err := r.db.Model(&model.ProductPackageAttributeGroup{}).Create(productPackage.ProductPackageAttributeGroups).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 创建product_package_attribute表数据
	for _, attributeGroup := range productPackage.ProductPackageAttributeGroups {
		if err := r.db.Model(&model.ProductPackageAttribute{}).Create(attributeGroup.ProductPackageAttributes).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	// 创建product_bom表数据
	if err := r.db.Model(&model.ProductBom{}).Create(productPackage.ProductBoms).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 创建related_material表数据
	for _, bom := range productPackage.ProductBoms {
		if err := r.db.Model(&model.RelatedMaterial{}).Create(bom.FlavorMaterials).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
