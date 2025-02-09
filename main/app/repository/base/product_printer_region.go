package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPrinterRegionRepo 商品打印（档口）
type IProductPrinterRegionRepo interface {
	GetProductPrinterRegionList() ([]model.ProductPrinterRegion, error)
	UpdateProductPrinterRegion(id uint, productPrinterRegion model.ProductPrinterRegion) error
	CreateProductPrinterRegion(productPrinterRegion model.ProductPrinterRegion) (uint, error)
	DeleteProductPrinterRegion(id uint) error
}

func NewProductPrinterRegionRepo(db *gorm.DB) IProductPrinterRegionRepo {
	return NewProductPrinterRegionRepoImpl(db)
}

// NewProductPrinterRepoImpl 创建新的商品打印（档口）仓库实现
func NewProductPrinterRegionRepoImpl(db *gorm.DB) *ProductPrinterRegionRepoImpl {
	return &ProductPrinterRegionRepoImpl{db: db}
}

type ProductPrinterRegionRepoImpl struct {
	db *gorm.DB
}

// GetProductPrinterRegionList 获取商品打印（档口）区域列表，排除逻辑删除的商品打印（档口）区域
func (r *ProductPrinterRegionRepoImpl) GetProductPrinterRegionList() ([]model.ProductPrinterRegion, error) {
	var productPrinterRegions []model.ProductPrinterRegion
	err := r.db.Model(&model.ProductPrinterRegion{}).Where("delete_time = ?", 0).Find(&productPrinterRegions).Error
	return productPrinterRegions, err
}

// UpdateProductPrinterRegion 更新商品打印（档口）区域
func (r *ProductPrinterRegionRepoImpl) UpdateProductPrinterRegion(uuid uint, productPrinterRegion model.ProductPrinterRegion) error {
	return r.db.Model(&model.ProductPrinterRegion{}).Where("uuid = ?", uuid).Updates(productPrinterRegion).Error
}

// CreateProductPrinterRegion 创建商品打印（档口）区域
func (r *ProductPrinterRegionRepoImpl) CreateProductPrinterRegion(productPrinterRegion model.ProductPrinterRegion) (uint, error) {
	return productPrinterRegion.ID, r.db.Create(&productPrinterRegion).Error
}

// DeleteProductPrinterRegion 软删除商品打印（档口）区域
func (r *ProductPrinterRegionRepoImpl) DeleteProductPrinterRegion(uuid uint) error {
	return r.db.Model(&model.ProductPrinterRegion{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
