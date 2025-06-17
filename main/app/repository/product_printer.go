package repository

import (
	"gorm.io/gorm"
	"slices"
	"ttpos-server-go/app/model"
)

type IProductPrinterRepo interface {
	WhereStatus(status int) DBOption
	WhereProductPrinterUuid(uuid uint64) DBOption
	WidthPrintMode(widthPrintMode int) DBOption

	GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error)   // 获取商品打印
	GetProductPackageUuids(opts ...DBOption) ([]uint64, error)             // 获取指定商品打印关联的商品Uuid
	GetProductionSaleBillUuid(productPrinterUuid uint64) ([]uint64, error) // 获取sale_bill_uuid
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
	db := r.db.Model(&model.ProductPrinter{}).Scopes(NotDeleted)

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
	err := db.Select("product_package_uuid").
		Where("product_package_uuid not in (?)", r.db.Model(&model.ProductPackage{}).Scopes(NotDeleted).Where("is_show_kitchen = 0").Select("uuid")).
		Pluck("product_package_uuid", &uuids).Error
	if err != nil {
		return nil, err
	}
	return uuids, err
}

func (r *productPrinterRepo) GetProductionSaleBillUuid(productPrinterUuid uint64) ([]uint64, error) {
	var deskRegionUuids []uint64
	r.db.Model(&model.ProductPrinterRegion{}).Scopes(NotDeleted).Where("product_printer_uuid = ?", productPrinterUuid).Pluck("desk_region_uuid", &deskRegionUuids)
	var deskUuids []uint64
	r.db.Model(&model.Desk{}).Scopes(NotDeleted).Where("region_uuid in (?)", deskRegionUuids).Pluck("uuid", &deskUuids)
	if slices.Contains(deskRegionUuids, 0) {
		deskUuids = append(deskUuids, 0)
	}
	var uuids []uint64
	err := r.db.Model(&model.SaleBill{}).Scopes(NotDeleted).Where("desk_uuid in (?)", deskUuids).Pluck("uuid", &uuids).Error
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

func (r *productPrinterRepo) WidthPrintMode(printMode int) DBOption {
	if printMode == -1 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("print_mode = ?", printMode)
	}
}
