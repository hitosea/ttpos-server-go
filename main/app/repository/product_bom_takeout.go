package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomTakeoutRepo interface {
	IProductBomTakeoutQueryRepo
	CreateProductBomTakeout(productBomTakeout *model.ProductBomTakeout) error
	UpdateProductBomTakeout(data map[string]any, opts ...DBOption) error
	DestroyProductBomTakeout(opts ...DBOption) error
	ForceDestroyProductBomTakeout(opts ...DBOption) error
	DeleteProductBomTakeout(opts ...DBOption) error // 物理删除
}

type IProductBomTakeoutQueryRepo interface {
	GetProductBomTakeout(opts ...DBOption) (*model.ProductBomTakeout, error)
	GetProductBomTakeoutList(opts ...DBOption) ([]*model.ProductBomTakeout, error)

	// 预加载
	WithProductBom(opts ...DBOption) DBOption
	WithProductBomProductFlavor(opts ...DBOption) DBOption
}

type productBomTakeoutRepoImpl struct {
	db *gorm.DB
}

func NewProductBomTakeoutRepo(db *gorm.DB) IProductBomTakeoutRepo {
	return &productBomTakeoutRepoImpl{db: db}
}

func (r *productBomTakeoutRepoImpl) CreateProductBomTakeout(productBomTakeout *model.ProductBomTakeout) error {
	return r.db.Create(productBomTakeout).Error
}

func (r *productBomTakeoutRepoImpl) UpdateProductBomTakeout(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.ProductBomTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Updates(data).Error
}

func (r *productBomTakeoutRepoImpl) DestroyProductBomTakeout(opts ...DBOption) error {
	db := r.db.Model(&model.ProductBomTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

func (r *productBomTakeoutRepoImpl) ForceDestroyProductBomTakeout(opts ...DBOption) error {
	db := r.db.Model(&model.ProductBomTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.ProductBomTakeout{}).Error
}

// DeleteProductBomTakeout 物理删除外卖规格价格记录
func (r *productBomTakeoutRepoImpl) DeleteProductBomTakeout(opts ...DBOption) error {
	db := r.db.Unscoped() // Unscoped 允许物理删除

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.ProductBomTakeout{}).Error
}

func (r *productBomTakeoutRepoImpl) GetProductBomTakeout(opts ...DBOption) (*model.ProductBomTakeout, error) {
	var productBomTakeout model.ProductBomTakeout
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBomTakeout)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBomTakeout, nil
}

func (r *productBomTakeoutRepoImpl) GetProductBomTakeoutList(opts ...DBOption) ([]*model.ProductBomTakeout, error) {
	var list []*model.ProductBomTakeout
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

// WithProductBom 预加载店内商品BOM
func (r *productBomTakeoutRepoImpl) WithProductBom(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBom", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductBomProductFlavor 预加载店内商品BOM的规格信息
func (r *productBomTakeoutRepoImpl) WithProductBomProductFlavor(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBom.ProductFlavor", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		}).Preload("ProductBom.ProductFlavor.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}
