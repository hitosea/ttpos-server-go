package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageRepo interface {
	IProductPackageQueryRepo
	CreateProductPackage(productPackage *model.ProductPackage) error
	UpdateProductPackage(data map[string]any, opts ...DBOption) error
	AddActualSaleNum(productPackageUuid uint64, saleNum float64) error
	SubActualSaleNum(productPackageUuid uint64, saleNum float64) error
}

type IProductPackageQueryRepo interface {
	GetProductPackage(opts ...DBOption) (*model.ProductPackage, error)
	GetProductPackageList(opts ...DBOption) ([]*model.ProductPackage, error)
	GetProductPackageBoms(productPackageUuid uint64) (*model.ProductPackage, error) // 获取商品包的库存信息
	GetProductPackageBaseInfoByBomUuid(flavorBomUuid uint64) (*model.ProductBom, error)
	GetProductPackageListByUuids(uuids []uint64) ([]*model.ProductPackage, error)
	GetProductPackageBatchTagCount() ([]uint64, error) // 获取分批商品数量
	SetProductPackageBatch(uuids []uint64) error       // 将is_batch设置为1
}

type productPackageRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageRepo(db *gorm.DB) IProductPackageRepo {
	return &productPackageRepoImpl{db: db}
}

func (r *productPackageRepoImpl) GetProductPackageList(opts ...DBOption) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	db := r.db

	db = db.Model(&model.ProductPackage{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productPackages)
	if result.Error != nil {
		return nil, result.Error
	}

	return productPackages, nil
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

// GetProductPackageBoms 查询商品包库存信息
func (r *productPackageRepoImpl) GetProductPackageBoms(productPackageUuid uint64) (*model.ProductPackage, error) {
	productPackage, err := r.GetProductPackage(
		CommonRepo.WhereByUuid(productPackageUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductBoms",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackage, nil
}

func (r *productPackageRepoImpl) GetProductPackageBaseInfoByBomUuid(flavorBomUuid uint64) (*model.ProductBom, error) {
	// 先查询出ProductBom
	productBomRepo := NewProductBomRepo(r.db)
	productBom, err := productBomRepo.GetProductBom(
		CommonRepo.WhereByUuid(flavorBomUuid),
		CommonRepo.Preload(WithPreload{Query: "ProductPackage.TakeoutTax"}),
		CommonRepo.Preload(
			WithPreload{Query: "ProductPackage.MultiLanguageName"},
			WithPreload{Query: "ProductPackage.ProductCategory"},
			WithPreload{Query: "ProductPackage.DineTax"},
			WithPreload{Query: "ProductFlavor.MultiLanguageName"},
		),
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
	multiLanguageNameRepo := NewMultiLanguageNameRepo(r.db)
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

// UpdateProductPackage 更新产品包
func (r *productPackageRepoImpl) UpdateProductPackage(data map[string]any, opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductPackage{}).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageRepoImpl) AddActualSaleNum(productPackageUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num + ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageRepoImpl) SubActualSaleNum(productPackageUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num - ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// GetProductPackageBatchTagCount 获取分批商品数量
func (r *productPackageRepoImpl) GetProductPackageBatchTagCount() ([]uint64, error) {
	var productPackageBatchTagCounts []uint64
	err := r.db.Model(&model.ProductPackage{}).Where("is_batch <> ?", 0).Where("delete_time = ?", 0).Select("uuid").Scan(&productPackageBatchTagCounts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageBatchTagCounts, nil
}

// 根据uuids将is_batch设置为1
func (r *productPackageRepoImpl) SetProductPackageBatch(uuids []uint64) error {
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ?", uuids).Updates(map[string]any{"is_batch": 1}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
