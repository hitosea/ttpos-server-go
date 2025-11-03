package repository

import (
	"strings"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductLabelRepo 商品标签仓库接口
type IProductLabelRepo interface {
	GetProductLabelList(opts ...DBOption) ([]model.ProductLabel, error)
	GetProductLabel(opts ...DBOption) (*model.ProductLabel, error)
	GetProductLabelByUuid(uuid uint64) (*model.ProductLabel, error)
	CreateProductLabel(productLabel model.ProductLabel) (uint64, error)
	UpdateProductLabel(productLabel model.ProductLabel) error
	DeleteProductLabel(uuid uint64) error
	ClearProductPackageLabelRelation(labelUuid uint64) error                                // 清除商品包与标签的关联关系
	UpdateProductPackageLabelRelation(productPackageUuids []uint64, labelUuid uint64) error // 建立商品包与标签的关联关系

	// 查询选项
	WhereUuid(uuid uint64) DBOption
	WithProductPackages() DBOption                    // 预加载关联的商品列表
	WithProductPackageCount() DBOption                // 统计关联的商品数量
	OrderByCreateTime(order string) DBOption          // 按创建时间排序
	OrderByProductPackageCount(order string) DBOption // 按关联商品数量排序
}

// NewProductLabelRepo 创建新的商品标签仓库
func NewProductLabelRepo(db *gorm.DB) IProductLabelRepo {
	return NewProductLabelRepoImpl(db)
}

// NewProductLabelRepoImpl 创建新的商品标签仓库实现
func NewProductLabelRepoImpl(db *gorm.DB) *ProductLabelRepoImpl {
	return &ProductLabelRepoImpl{db: db}
}

type ProductLabelRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductLabelList 获取商品标签列表
func (r *ProductLabelRepoImpl) GetProductLabelList(opts ...DBOption) ([]model.ProductLabel, error) {
	var productLabels []model.ProductLabel
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&productLabels).Error
	return productLabels, errors.WithMessage(err)
}

// GetProductLabel 获取商品标签
func (r *ProductLabelRepoImpl) GetProductLabel(opts ...DBOption) (*model.ProductLabel, error) {
	var productLabel model.ProductLabel
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&productLabel).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productLabel, nil
}

// GetProductLabelByUuid 根据 UUID 获取商品标签
func (r *ProductLabelRepoImpl) GetProductLabelByUuid(uuid uint64) (*model.ProductLabel, error) {
	label, err := r.GetProductLabel(r.WhereUuid(uuid), CommonRepo.WhereBySoftDelete())
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return nil, errors.WithMessage(err)
		}
	}
	return label, nil
}

// CreateProductLabel 创建商品标签
func (r *ProductLabelRepoImpl) CreateProductLabel(productLabel model.ProductLabel) (uint64, error) {
	err := r.db.Create(&productLabel).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return productLabel.Uuid, nil
}

// UpdateProductLabel 更新商品标签
func (r *ProductLabelRepoImpl) UpdateProductLabel(productLabel model.ProductLabel) error {
	err := r.db.Model(&model.ProductLabel{}).Where("uuid = ?", productLabel.Uuid).Updates(&productLabel).Error
	return errors.WithMessage(err)
}

// DeleteProductLabel 软删除商品标签
func (r *ProductLabelRepoImpl) DeleteProductLabel(uuid uint64) error {
	err := r.db.Model(&model.ProductLabel{}).
		Where("uuid = ?", uuid).
		Update("delete_time", time.Now().Unix()).Error
	return errors.WithMessage(err)
}

// ClearProductPackageLabelRelation 清除商品包与标签的关联关系
func (r *ProductLabelRepoImpl) ClearProductPackageLabelRelation(labelUuid uint64) error {
	err := r.db.Model(&model.ProductPackage{}).
		Where("product_label_uuid = ?", labelUuid).
		Update("product_label_uuid", 0).Error
	return errors.WithMessage(err)
}

// UpdateProductPackageLabelRelation 建立商品包与标签的关联关系
func (r *ProductLabelRepoImpl) UpdateProductPackageLabelRelation(productPackageUuids []uint64, labelUuid uint64) error {
	err := r.db.Model(&model.ProductPackage{}).
		Where("uuid IN ?", productPackageUuids).
		Update("product_label_uuid", labelUuid).Error
	return errors.WithMessage(err)
}

// WhereUuid 按 UUID 查询
func (r *ProductLabelRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WithProductPackages 预加载关联的商品列表
func (r *ProductLabelRepoImpl) WithProductPackages() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackages", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = 0")
		}).Preload("ProductPackages.MultiLanguageName")
	}
}

// WithProductPackageCount 统计关联的商品数量
func (r *ProductLabelRepoImpl) WithProductPackageCount() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select("ttpos_product_label.*, COUNT(ttpos_product_package.uuid) as product_package_count").
			Joins("LEFT JOIN ttpos_product_package ON ttpos_product_label.uuid = ttpos_product_package.product_label_uuid AND ttpos_product_package.delete_time = 0").
			Group("ttpos_product_label.uuid")
	}
}

// OrderByCreateTime 按创建时间排序
func (r *ProductLabelRepoImpl) OrderByCreateTime(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if order == "" {
			order = "DESC"
		}
		return db.Order("create_time " + order)
	}
}

// OrderByProductPackageCount 按关联商品数量排序
func (r *ProductLabelRepoImpl) OrderByProductPackageCount(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if order == "" {
			order = "DESC"
		}
		return db.Order("product_package_count " + order)
	}
}
