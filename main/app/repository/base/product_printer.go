package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPrinterRepo 商品打印（档口）
type IProductPrinterRepo interface {
	GetProductPrinterList() ([]model.ProductPrinter, error)
	UpdateProductPrinter(id uint, productPrinter model.ProductPrinter) error
	CreateProductPrinter(productPrinter model.ProductPrinter) (uint, error)
	DeleteProductPrinter(id uint) error
}

func NewProductPrinterRepo(db *gorm.DB) IProductPrinterRepo {
	return NewProductPrinterRepoImpl(db)
}

// NewProductPrinterRepoImpl 创建新的商品打印（档口）仓库实现
func NewProductPrinterRepoImpl(db *gorm.DB) *ProductPrinterRepoImpl {
	return &ProductPrinterRepoImpl{db: db}
}

type ProductPrinterRepoImpl struct {
	db *gorm.DB
}

// GetProductPrinterList 获取商品打印（档口）列表，排除逻辑删除的商品打印（档口）
func (r *ProductPrinterRepoImpl) GetProductPrinterList() ([]model.ProductPrinter, error) {
	var productPrinters []model.ProductPrinter
	err := r.db.Model(&model.ProductPrinter{}).Where("delete_time = ?", 0).Find(&productPrinters).Error
	return productPrinters, err
}

// UpdateProductPrinter 更新商品打印（档口）
func (r *ProductPrinterRepoImpl) UpdateProductPrinter(uuid uint, productPrinter model.ProductPrinter) error {
	return r.db.Model(&model.ProductPrinter{}).Where("uuid = ?", uuid).Updates(productPrinter).Error
}

// CreateProductPrinter 创建商品打印（档口）
func (r *ProductPrinterRepoImpl) CreateProductPrinter(productPrinter model.ProductPrinter) (uint, error) {
	return productPrinter.ID, r.db.Create(&productPrinter).Error
}

// DeleteProductPrinter 软删除商品打印（档口）
func (r *ProductPrinterRepoImpl) DeleteProductPrinter(uuid uint) error {
	return r.db.Model(&model.ProductPrinter{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
