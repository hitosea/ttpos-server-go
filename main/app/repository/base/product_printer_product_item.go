package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPrinterProductItemRepo 商品打印（档口）产品关联表
type IProductPrinterProductItemRepo interface {
	GetProductPrinterProductItemList() ([]model.ProductPrinterProductItem, error)
	UpdateProductPrinterProductItem(id uint, productPrinterProductItem model.ProductPrinterProductItem) error
	CreateProductPrinterProductItem(productPrinterProductItem model.ProductPrinterProductItem) (uint, error)
	DeleteProductPrinterProductItem(id uint) error
}

func NewProductPrinterProductItemRepo(db *gorm.DB) IProductPrinterProductItemRepo {
	return NewProductPrinterProductItemRepoImpl(db)
}

// NewProductPrinterProductItemRepoImpl 创建新的商品打印（档口）产品关联表仓库实现
func NewProductPrinterProductItemRepoImpl(db *gorm.DB) *ProductPrinterProductItemRepoImpl {
	return &ProductPrinterProductItemRepoImpl{db: db}
}

type ProductPrinterProductItemRepoImpl struct {
	db *gorm.DB
}

// GetProductPrinterProductItemList 获取商品打印（档口）产品关联列表，排除逻辑删除的商品打印（档口）产品关联
func (r *ProductPrinterProductItemRepoImpl) GetProductPrinterProductItemList() ([]model.ProductPrinterProductItem, error) {
	var productPrinterProductItems []model.ProductPrinterProductItem
	err := r.db.Model(&model.ProductPrinterProductItem{}).Where("delete_time = ?", 0).Find(&productPrinterProductItems).Error
	return productPrinterProductItems, errors.WithMessage(err)
}

// UpdateProductPrinterProductItem 更新商品打印（档口）产品关联
func (r *ProductPrinterProductItemRepoImpl) UpdateProductPrinterProductItem(uuid uint, productPrinterProductItem model.ProductPrinterProductItem) error {
	return r.db.Model(&model.ProductPrinterProductItem{}).Where("uuid = ?", uuid).Updates(productPrinterProductItem).Error
}

// CreateProductPrinterProductItem 创建商品打印（档口）产品关联
func (r *ProductPrinterProductItemRepoImpl) CreateProductPrinterProductItem(productPrinterProductItem model.ProductPrinterProductItem) (uint, error) {
	return productPrinterProductItem.ID, r.db.Create(&productPrinterProductItem).Error
}

// DeleteProductPrinterProductItem 软删除商品打印（档口）产品关联
func (r *ProductPrinterProductItemRepoImpl) DeleteProductPrinterProductItem(uuid uint) error {
	return r.db.Model(&model.ProductPrinterProductItem{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
