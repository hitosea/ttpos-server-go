package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPrinterItemRepo 商品打印（档口）打印机关联表
type IProductPrinterItemRepo interface {
	GetProductPrinterItemList() ([]model.ProductPrinterItem, error)
	UpdateProductPrinterItem(id uint, productPrinterItem model.ProductPrinterItem) error
	CreateProductPrinterItem(productPrinterItem model.ProductPrinterItem) (uint, error)
	DeleteProductPrinterItem(id uint) error
}

func NewProductPrinterItemRepo(db *gorm.DB) IProductPrinterItemRepo {
	return NewProductPrinterItemRepoImpl(db)
}

// NewProductPrinterItemRepoImpl 创建新的商品打印（档口）打印机关联表仓库实现
func NewProductPrinterItemRepoImpl(db *gorm.DB) *ProductPrinterItemRepoImpl {
	return &ProductPrinterItemRepoImpl{db: db}
}

type ProductPrinterItemRepoImpl struct {
	db *gorm.DB
}

// GetProductPrinterItemList 获取商品打印（档口）打印机关联列表，排除逻辑删除的商品打印（档口）打印机关联
func (r *ProductPrinterItemRepoImpl) GetProductPrinterItemList() ([]model.ProductPrinterItem, error) {
	var productPrinterItems []model.ProductPrinterItem
	err := r.db.Model(&model.ProductPrinterItem{}).Where("delete_time = ?", 0).Find(&productPrinterItems).Error
	return productPrinterItems, err
}

// UpdateProductPrinterItem 更新商品打印（档口）打印机关联
func (r *ProductPrinterItemRepoImpl) UpdateProductPrinterItem(uuid uint, productPrinterItem model.ProductPrinterItem) error {
	return r.db.Model(&model.ProductPrinterItem{}).Where("uuid = ?", uuid).Updates(productPrinterItem).Error
}

// CreateProductPrinterItem 创建商品打印（档口）打印机关联
func (r *ProductPrinterItemRepoImpl) CreateProductPrinterItem(productPrinterItem model.ProductPrinterItem) (uint, error) {
	return productPrinterItem.ID, r.db.Create(&productPrinterItem).Error
}

// DeleteProductPrinterItem 软删除商品打印（档口）打印机关联
func (r *ProductPrinterItemRepoImpl) DeleteProductPrinterItem(uuid uint) error {
	return r.db.Model(&model.ProductPrinterItem{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
