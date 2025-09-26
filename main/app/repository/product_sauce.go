package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductSauceRepo interface {
	GetProductSauce(opts ...DBOption) (*model.ProductSauce, error)
	CreateProductSauce(productSauce *model.ProductSauce) error
	GetSauceByUuid(uuid uint64) (*model.ProductSauce, error)
	UpdateProductBomCard(uuid uint64, productBomCardUuid uint64) error
	GetSauceByErpCode(erpCode string) (*model.ProductSauce, error)
	GetSauceByProductBomCardUuid(productBomCardUuid uint64) (*model.ProductSauce, error)
}

type productSauceRepoImpl struct {
	db *gorm.DB
}

func NewProductSauceRepo(db *gorm.DB) IProductSauceRepo {
	return &productSauceRepoImpl{db: db}
}

func (r *productSauceRepoImpl) GetProductSauce(opts ...DBOption) (*model.ProductSauce, error) {
	var productSauce model.ProductSauce
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productSauce)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return &productSauce, nil
}

func (r *productSauceRepoImpl) CreateProductSauce(productSauce *model.ProductSauce) error {
	// 创建product_sauce表数据
	sauce := *productSauce
	sauce.SetNil()
	if err := r.db.Model(&model.ProductSauce{}).Create(&sauce).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 创建multi_language_name表数据
	if _, err := NewMultiLanguageNameRepo(r.db).CreateMultiLanguageName(productSauce.MultiLanguageName); err != nil {
		return errors.WithMessage(err)
	}

	// 创建product_sauce_material表数据
	productSauceMaterials := make([]model.RelatedMaterial, 0)
	for _, material := range productSauce.SauceMaterials {
		productSauceMaterials = append(productSauceMaterials, *material)
	}
	if err := NewRelatedMaterialRepo(r.db).CreateRelatedMaterials(productSauceMaterials); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productSauceRepoImpl) GetSauceByUuid(uuid uint64) (*model.ProductSauce, error) {
	sauce, err := r.GetProductSauce(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return sauce, nil
}

func (r *productSauceRepoImpl) UpdateProductBomCard(uuid uint64, productBomCardUuid uint64) error {
	return r.db.Model(&model.ProductSauce{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"product_bom_card_uuid": productBomCardUuid,
	}).Error
}

func (r *productSauceRepoImpl) GetSauceByErpCode(erpCode string) (*model.ProductSauce, error) {
	var productSauce model.ProductSauce
	if err := r.db.Model(&model.ProductSauce{}).Where("erp_code = ?", erpCode).Where("delete_time = 0").
		Preload("MultiLanguageName").
		First(&productSauce).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productSauce, nil
}

func (r *productSauceRepoImpl) GetSauceByProductBomCardUuid(productBomCardUuid uint64) (*model.ProductSauce, error) {
	var productSauce model.ProductSauce
	if err := r.db.Model(&model.ProductSauce{}).Where("product_bom_card_uuid = ?", productBomCardUuid).Where("delete_time = 0").
		Preload("MultiLanguageName").
		First(&productSauce).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productSauce, nil
}
