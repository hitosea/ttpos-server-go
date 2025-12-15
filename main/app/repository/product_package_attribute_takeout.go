package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageAttributeTakeoutRepo interface {
	IProductPackageAttributeTakeoutQueryRepo
	CreateProductPackageAttributeTakeout(productPackageAttributeTakeout *model.ProductPackageAttributeTakeout) error
	UpdateProductPackageAttributeTakeout(data map[string]any, opts ...DBOption) error
	DestroyProductPackageAttributeTakeout(opts ...DBOption) error
}

type IProductPackageAttributeTakeoutQueryRepo interface {
	GetProductPackageAttributeTakeout(opts ...DBOption) (*model.ProductPackageAttributeTakeout, error)
	GetProductPackageAttributeTakeoutList(opts ...DBOption) ([]*model.ProductPackageAttributeTakeout, error)

	// 预加载
	WithProductPackageAttribute(opts ...DBOption) DBOption
	WithProductPackageAttributeAttribute(opts ...DBOption) DBOption
}

type productPackageAttributeTakeoutRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageAttributeTakeoutRepo(db *gorm.DB) IProductPackageAttributeTakeoutRepo {
	return &productPackageAttributeTakeoutRepoImpl{db: db}
}

func (r *productPackageAttributeTakeoutRepoImpl) CreateProductPackageAttributeTakeout(productPackageAttributeTakeout *model.ProductPackageAttributeTakeout) error {
	return r.db.Create(productPackageAttributeTakeout).Error
}

func (r *productPackageAttributeTakeoutRepoImpl) UpdateProductPackageAttributeTakeout(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.ProductPackageAttributeTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Updates(data).Error
}

func (r *productPackageAttributeTakeoutRepoImpl) DestroyProductPackageAttributeTakeout(opts ...DBOption) error {
	db := r.db.Model(&model.ProductPackageAttributeTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

func (r *productPackageAttributeTakeoutRepoImpl) GetProductPackageAttributeTakeout(opts ...DBOption) (*model.ProductPackageAttributeTakeout, error) {
	var productPackageAttributeTakeout model.ProductPackageAttributeTakeout
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackageAttributeTakeout)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackageAttributeTakeout, nil
}

func (r *productPackageAttributeTakeoutRepoImpl) GetProductPackageAttributeTakeoutList(opts ...DBOption) ([]*model.ProductPackageAttributeTakeout, error) {
	var list []*model.ProductPackageAttributeTakeout
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}

	return list, nil
}

// WithProductPackageAttribute 预加载店内商品属性
func (r *productPackageAttributeTakeoutRepoImpl) WithProductPackageAttribute(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttribute", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageAttributeAttribute 预加载店内商品属性的属性信息
func (r *productPackageAttributeTakeoutRepoImpl) WithProductPackageAttributeAttribute(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttribute.Attribute", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		}).Preload("ProductPackageAttribute.Attribute.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}
