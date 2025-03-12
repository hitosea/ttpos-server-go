package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IProductPrinterRepo interface {
	WhereStatus(status int) DBOption
	WhereProductPrinterUuid(uuid uint64) DBOption

	GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error) // 获取商品打印
	GetProductPackageUuids(opts ...DBOption) ([]uint64, error)           // 获取指定商品打印关联的商品Uuid
}

func NewProductPrinterRepo(db *gorm.DB) IProductPrinterRepo {
	return NewProductPrinterRepoImpl(db)
}

type productPrinterRepo struct {
	db *gorm.DB
}

func NewProductPrinterRepoImpl(db *gorm.DB) IProductPrinterRepo {
	return &productPrinterRepo{db: db}
}

func (r *productPrinterRepo) GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error) {
	var printers []model.ProductPrinter
	db := r.db.Model(&model.ProductPrinter{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Order("create_time desc").Find(&printers).Error
	return printers, err
}

func (r *productPrinterRepo) GetProductPackageUuids(opts ...DBOption) ([]uint64, error) {
	var uuids []uint64
	db := r.db.Model(&model.ProductPrinterProductItem{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("product_package_uuid").Pluck("product_package_uuid", &uuids).Error
	if err != nil {
		return nil, err
	}
	return uuids, err
}

func (r *productPrinterRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}
func (r *productPrinterRepo) WhereProductPrinterUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_printer_uuid = ?", uuid)
	}
}
