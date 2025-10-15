package repository

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPrinterRepo interface {
	WhereStatus(status int) DBOption
	WhereProductPrinterUuid(uuid uint64) DBOption
	WidthPrintMode(widthPrintMode int) DBOption
	WhereSaleBillIsKitchenConfirm(isKitchenConfirm int) DBOption // 厨显端是否确认退菜整单
	WhereSaleBillNotDeletedOrIsNotCanceled() DBOption            // 未被删除的，未整单取消的

	GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error)                                                // 获取商品打印
	GetProductPrintersByProductPackageUuid(productPackageUuid uint64, opts ...DBOption) ([]model.ProductPrinter, error) // 获取指定商品打印关联的商品Uuid
	GetProductPackageUuids(opts ...DBOption) ([]uint64, error)                                                          // 获取指定商品打印关联的商品Uuid
	GetProductionSaleBillUuid(productPrinterUuid uint64, opt DBOption) ([]uint64, error)                                // 获取sale_bill_uuid
	CreateProductPackagePrinter(productPackageUuid uint64, productPrinterUuids []uint64) error                          // 创建商品包关联打印机
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

// 获取指定商品打印关联的商品Uuid
func (r *productPrinterRepo) GetProductPrintersByProductPackageUuid(productPackageUuid uint64, opts ...DBOption) ([]model.ProductPrinter, error) {
	// 第一步：从关联表中查询product_printer_uuid列表
	var productPrinterUuids []uint64
	err := r.db.Model(&model.ProductPrinterProductItem{}).
		Scopes(NotDeleted).
		Where("product_package_uuid = ?", productPackageUuid).
		Pluck("product_printer_uuid", &productPrinterUuids).Error
	if err != nil {
		return nil, err
	}

	// 如果没有关联的打印机，直接返回空列表
	if len(productPrinterUuids) == 0 {
		return []model.ProductPrinter{}, nil
	}

	// 第二步：根据product_printer_uuid列表查询ProductPrinter
	var printers []model.ProductPrinter
	db := r.db.Model(&model.ProductPrinter{}).
		Scopes(NotDeleted).
		Where("uuid IN (?)", productPrinterUuids)

	for _, opt := range opts {
		db = opt(db)
	}

	err = db.Order("create_time desc").Find(&printers).Error
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

func (r *productPrinterRepo) GetProductionSaleBillUuid(productPrinterUuid uint64, opt DBOption) ([]uint64, error) {
	var deskRegionUuids []uint64
	r.db.Model(&model.ProductPrinterRegion{}).Scopes(NotDeleted).Where("product_printer_uuid = ?", productPrinterUuid).Pluck("desk_region_uuid", &deskRegionUuids)
	var deskUuids []uint64
	if len(deskRegionUuids) != 0 {
		r.db.Model(&model.Desk{}).Scopes(NotDeleted).Where("region_uuid in (?)", deskRegionUuids).Pluck("uuid", &deskUuids)
		if slices.Contains(deskRegionUuids, 0) {
			deskUuids = append(deskUuids, 0)
		}
	}
	var uuids []uint64
	query := r.db.Model(&model.SaleBill{})
	query = opt(query)
	if len(deskUuids) != 0 {
		query = query.Where("desk_uuid in (?)", deskUuids)
	}
	// 排除仅有is_batch=1（分批商品）且batch_time=0的sale_bill_uuid
	query = query.Where("uuid in (?)", r.db.Model(&model.ProductionOrderProduct{}).Select("sale_bill_uuid").Where("is_batch = 0 OR batch_time > 0").Group("sale_bill_uuid"))
	err := query.Pluck("uuid", &uuids).Error
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
	if printMode == -1 || printMode == -2 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("print_mode = ?", printMode)
	}
}

func (r *productPrinterRepo) WhereSaleBillIsKitchenConfirm(isKitchenConfirm int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_kitchen_confirm = ?", isKitchenConfirm)
	}
}

// 未被删除的，未整单取消的
func (r *productPrinterRepo) WhereSaleBillNotDeletedOrIsNotCanceled() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ? or status <> ?", constant.NotDeleted, constant.SaleBillStatusCanceled)
	}
}

// 创建商品包关联打印机（先删除旧关联，再批量插入新关联）
func (r *productPrinterRepo) CreateProductPackagePrinter(productPackageUuid uint64, productPrinterUuids []uint64) error {
	// 使用事务确保数据一致性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 第一步：删除该商品包的所有旧关联
		err := tx.Where("product_package_uuid = ?", productPackageUuid).Delete(&model.ProductPrinterProductItem{}).Error
		if err != nil {
			return err
		}

		// 如果没有新的打印机关联，直接返回
		if len(productPrinterUuids) == 0 {
			return nil
		}

		// 第二步：批量插入新的关联记录
		var items []model.ProductPrinterProductItem
		for _, productPrinterUuid := range productPrinterUuids {
			items = append(items, model.ProductPrinterProductItem{
				ProductPackageUuid: productPackageUuid,
				ProductPrinterUuid: productPrinterUuid,
			})
		}

		// 批量创建
		return tx.Create(&items).Error
	})
}
