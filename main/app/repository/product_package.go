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
	productPkg := *productPackage
	productPkg.SetNil()
	if err := r.db.Model(&model.ProductPackage{}).Create(&productPkg).Error; err != nil {
		return errors.WithMessage(err)
	}

	// 创建multi_language_name表数据
	multiLanguageNameRepo := NewMultiLanguageNameRepoImpl(r.db)
	if _, err := multiLanguageNameRepo.CreateMultiLanguageName(productPackage.MultiLanguageName); err != nil {
		return errors.WithMessage(err)
	}

	productPackageAttributeGroupRepo := NewProductPackageAttributeGroupRepo(r.db)
	if err := productPackageAttributeGroupRepo.CreateProductPackageAttributeGroups(productPackage.ProductPackageAttributeGroups); err != nil {
		return errors.WithMessage(err)
	}

	// 创建product_package_attribute表数据
	productPackageAttributeRepo := NewProductPackageAttributeRepo(r.db)
	for _, attributeGroup := range productPackage.ProductPackageAttributeGroups {
		if err := productPackageAttributeRepo.CreateProductPackageAttributes(attributeGroup.ProductPackageAttributes); err != nil {
			return errors.WithMessage(err)
		}
	}
	// 创建product_bom表数据
	productBomRepo := NewProductBomRepo(r.db)
	if err := productBomRepo.CreateProductBoms(productPackage.ProductBoms); err != nil {
		return errors.WithMessage(err)
	}

	// 创建related_material表数据
	relatedMaterialRepo := NewRelatedMaterialRepo(r.db)
	for _, bom := range productPackage.ProductBoms {
		if len(bom.FlavorMaterials) > 0 {
			relatedMaterials := make([]model.RelatedMaterial, 0)
			for _, material := range bom.FlavorMaterials {
				relatedMaterials = append(relatedMaterials, *material)
			}
			if err := relatedMaterialRepo.CreateRelatedMaterials(relatedMaterials); err != nil {
				return errors.WithMessage(err)
			}
		}
	}
	return nil
}
