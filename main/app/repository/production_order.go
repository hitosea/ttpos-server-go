package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IProductionOrderRepo interface {
	IProductionOrderQueryRepo
	CreateProductionOrder(order *model.ProductionOrder) error // 创建生产订单
	WhereProductStatus(status uint) DBOption                  // 生产商品状态
	WhereUuid(uuid uint64) DBOption                           // Uuid 条件
	WhereProductFinishedTime(finishedTime int64) DBOption     // 生产商品完成时间条件
	WhereProductMadeTime(madeTime int64) DBOption             // 生产商品制作时间条件
	WhereProductMakeStatus(finishStatus []uint) DBOption      // 制作状态
	WhereProductUuid(uuid uint64) DBOption                    // 生产商品Uuid条件
	WhereProductUuidIn(uuids []uint64) DBOption               // 生产商品Uuid条件列表
	WhereSaleOrderProductUuid(uuid uint64) DBOption           // 生产商品销售订单uuid条件
	WhereSaleBillUuidIn(uuids []uint64) DBOption              // 销售账单uuid条件
	WhereSaleBillUuid(uuid uint64) DBOption                   // 销售账单uuid条件
	WhereSource(source string) DBOption                       // 来源条件
	WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption  // 生产商品分类Uuid条件
	WhereProductNumGT0() DBOption                             // 送厨商品数量大于0
	WhereIsNotBatchOrBatchTimeGT0() DBOption                  // 非分批商品、或者分批已送厨商品

	WhereProductPackageInPrinter(productPrinterUuid uint64) DBOption                      // 商品在打印机关联中（子查询优化）
	WhereSaleBillInPrinterRegions(productPrinterUuid uint64, versionGte240 bool) DBOption // 销售账单在打印机关联区域中（子查询优化）

	SaleBillUuidOpt() DBOption                                                                                                                            // 历史上菜条件
	WithSaleOrderProductAll() DBOption                                                                                                                    // 关联销售订单商品
	WithProductCategory() DBOption                                                                                                                        // 关联商品分类
	WithProductCategoryMultiLanguageName() DBOption                                                                                                       // 关联商品分类多语言
	UpdateProduct(opts []DBOption, vars map[string]any) error                                                                                             // 更新送厨商品
	UpdateOrder(opts []DBOption, vars map[string]any) error                                                                                               // 更新送厨单
	IsProductionFinishedBySaleBillUuid(saleBillUuid uint64) (bool, error)                                                                                 // 检查销售账单下所有生产订单是否完成
	UpdateProductionOrderProductBatchTimeAndBatchTagUuid(saleBillUuid uint64, saleOrderProductUuids []uint64, batchTime int64, batchTagUuid uint64) error // 通过sale_bill_uuid和sale_order_product_uuid更新生产订单商品的batch_time、batch_tag_uuid
}

// IProductionOrderQueryRepo 生产订单查询仓库接口
type IProductionOrderQueryRepo interface {
	GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error)                                                                                              // 获取生产订单
	GetProduct(opts ...DBOption) (*model.ProductionOrderProduct, error)                                                                                               // 获取生产订单商品
	GetProductsByUuids(uuids []uint64, opts ...DBOption) ([]model.ProductionOrderProduct, error)                                                                      // 根据Uuid列表获取生产订单商品
	GetLimitedProducts(column string, pageNo, pageSize int, opts ...DBOption) ([]model.ProductionOrderProduct, int64, error)                                          // 分页获取账单ID、分类ID
	GetLimitedHistoryProducts(orderField string, opts ...DBOption) ([]model.ProductionOrderProduct, error)                                                            // 历史获取销售账单Uuid
	GetProducts(limit int, orderBy string, statusOpt DBOption, opts ...DBOption) (float64, []model.ProductionOrderProduct, error)                                     // 获取生产订单商品
	GetProductionOrderList(pageNo, pageSize int, productBomUuids []uint64, startTime int64, endTime int64, opts ...DBOption) ([]*model.ProductionOrderProduct, error) // 获取生产订单列表
	GetProductionOrderListCount(productBomUuids []uint64, startTime int64, endTime int64, opts ...DBOption) (int64, error)                                            // 获取生产订单列表数量

	GetProductsByPackageUuid(packageUuid uint64) ([]model.ProductionOrderProduct, error) // 根据套餐uuid获取套餐下所有子商品
}

type productionRepo struct {
	db *gorm.DB
}

func NewProductionRepo(db *gorm.DB) IProductionOrderRepo {
	return NewProductionOrderRepoImpl(db)
}

func NewProductionOrderRepoImpl(db *gorm.DB) IProductionOrderRepo {
	return &productionRepo{db: db}
}

func (r *productionRepo) GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error) {
	var productionOrder model.ProductionOrder
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Order("create_time desc").First(&productionOrder)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productionOrder, nil
}

func (r *productionRepo) GetProduct(opts ...DBOption) (*model.ProductionOrderProduct, error) {
	var product model.ProductionOrderProduct
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&product)
	if result.Error != nil {
		return nil, result.Error
	}

	return &product, nil
}

const (
	CreateTimeAsc    string = "create_time asc"
	FinishedTimeDesc string = "finished_time desc"
	MadeTimeDesc     string = "made_time desc"
)

func (r *productionRepo) GetProducts(limit int, orderBy string, statusOpt DBOption, opts ...DBOption) (float64, []model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	if statusOpt != nil {
		db = statusOpt(db).Session(&gorm.Session{})
	} else {
		db = db.Session(&gorm.Session{})
	}

	var total float64
	// 统计商品数量总和
	db.Select("IFNULL(SUM(num), 0) as total").Scan(&total)

	for _, opt := range opts {
		db = opt(db)
	}
	db.Preload("SaleBill").
		Preload("BatchTag.MultiLanguageName").
		Preload("SaleOrderProduct").
		Preload("SaleOrderProduct.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductBoms", NotDeleted).
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductAttributes", NotDeleted).
		Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute").
		Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName").Order(orderBy)
	if limit > 0 {
		db.Limit(limit)
	}
	result := db.Find(&productionOrderProducts)
	if result.Error != nil {
		return total, nil, result.Error
	}
	return total, productionOrderProducts, nil
}

// CreateProductionOrder 创建ProductionOrder记录及它管理的表记录
func (r *productionRepo) CreateProductionOrder(order *model.ProductionOrder) error {
	if err := r.db.Model(order).Create(order).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// GetLimitedProducts 分页获取账单ID、分类ID
func (r *productionRepo) GetLimitedProducts(column string, pageNo, pageSize int, opts ...DBOption) ([]model.ProductionOrderProduct, int64, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	var total int64
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted).Session(&gorm.Session{})
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	err := db.Select(fmt.Sprintf("count(distinct `%s`) as total", column)).Scan(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取列表
	err = db.Select("DISTINCT " + column).Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("create_time asc").Find(&productionOrderProducts).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return productionOrderProducts, total, nil
}

// GetLimitedHistoryProducts 历史获取销售账单Uuid
func (r *productionRepo) GetLimitedHistoryProducts(orderField string, opts ...DBOption) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Model(&model.ProductionOrderProduct{}).Select(fmt.Sprintf("sale_bill_uuid, MAX(%s) as finished_time", orderField)).Group("sale_bill_uuid").Order(orderField + " desc").Find(&productionOrderProducts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productionOrderProducts, nil
}

// SaleBillUuidOpt 历史上菜条件
func (r *productionRepo) SaleBillUuidOpt() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", r.db.Model(&model.ProductionOrderProduct{}).Select("DISTINCT sale_bill_uuid"))
	}
}

// WhereProductStatus 状态条件
func (r *productionRepo) WhereProductStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereUuid uuid条件
func (r *productionRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereProductFinishedTime 生产商品完成时间条件
func (r *productionRepo) WhereProductFinishedTime(finishedTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("finished_time > ?", finishedTime)
	}
}

// WhereProductMadeTime 生产商品制作时间条件
func (r *productionRepo) WhereProductMadeTime(madeTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("made_time > ?", madeTime)
	}
}

// WhereProductMakeStatus 制作状态条件
func (r *productionRepo) WhereProductMakeStatus(makeStatus []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("make_status in (?)", makeStatus)
	}
}

// WhereProductUuid 生产商品Uuid条件
func (r *productionRepo) WhereProductUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereProductUuidIn 生产商品Uuid条件列表
func (r *productionRepo) WhereProductUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid in (?)", uuids)
	}
}

// WhereSaleOrderProductUuid 生产商品销售订单商品Uuid条件
func (r *productionRepo) WhereSaleOrderProductUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_product_uuid = ?", uuid)
	}
}

// WhereSaleBillUuidIn 销售账单uuid条件
func (r *productionRepo) WhereSaleBillUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", uuids)
	}
}

// WhereProductPackageInPrinter 商品在打印机关联中（子查询优化，避免IN子句过长）
func (r *productionRepo) WhereProductPackageInPrinter(productPrinterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		prefix := config.Database.TablePrefix
		// 子查询：获取打印机关联的商品包UUID
		subQuery := r.db.Table(prefix+"product_printer_product_item").
			Select("product_package_uuid").
			Where("product_printer_uuid = ?", productPrinterUuid).
			Scopes(NotDeleted).
			Where("product_package_uuid NOT IN (?)",
				// 排除不在厨显显示的商品
				r.db.Table(prefix+"product_package").
					Select("uuid").
					Where("is_show_kitchen = ?", 0).
					Scopes(NotDeleted))

		return db.Where("product_package_uuid IN (?)", subQuery)
	}
}

// WhereSaleBillInPrinterRegions 销售账单在打印机关联区域中（子查询优化，避免IN子句过长）
func (r *productionRepo) WhereSaleBillInPrinterRegions(productPrinterUuid uint64, versionGte240 bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		prefix := config.Database.TablePrefix
		// 子查询1：获取打印机关联的区域UUID
		regionSubQuery := r.db.Table(prefix+"product_printer_region").
			Select("desk_region_uuid").
			Where("product_printer_uuid = ?", productPrinterUuid).
			Scopes(NotDeleted)

		// 子查询2：获取区域关联的桌台UUID
		deskSubQuery := r.db.Table(prefix+"desk").
			Select("uuid").
			Scopes(NotDeleted).
			Where("region_uuid IN (?) OR region_uuid = 0", regionSubQuery)

		// 子查询3：获取桌台关联的销售账单UUID
		billSubQuery := r.db.Table(prefix+"sale_bill").
			Select("uuid").
			Where("desk_uuid IN (?)", deskSubQuery)

		// 根据版本号决定过滤条件
		if versionGte240 {
			// 2.4.0 及以上版本：厨显端未确认退菜整单的账单
			billSubQuery = billSubQuery.Where("is_kitchen_confirm = ?", 0)
		} else {
			// 2.4.0 之前版本：未被删除的，未整单取消的
			billSubQuery = billSubQuery.Scopes(NotDeleted).Where("status <> ?", constant.SaleBillStatusCanceled)
		}
		return db.Where("sale_bill_uuid IN (?)", billSubQuery)
	}
}

// WhereSaleBillUuid 销售账单uuid条件
func (r *productionRepo) WhereSaleBillUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid = ?", uuid)
	}
}

// WhereSource 来源
func (r *productionRepo) WhereSource(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

// WhereProductFirstCategoryUuidIn 生产商品分类Uuid条件
func (r *productionRepo) WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("first_category_uuid in (?)", uuids)
	}
}

// WhereProductNumGT0 送厨商品数量大于0
func (r *productionRepo) WhereProductNumGT0() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("num > 0")
	}
}

// WithSaleBill 关联销售账单
func (r *productionRepo) WithSaleBill() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill")
	}
}

// WithSaleOrderProductAll 关联销售订单商品
func (r *productionRepo) WithSaleOrderProductAll() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProduct.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductBoms", NotDeleted).
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductAttributes", NotDeleted).
			Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute").
			Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName")
	}
}

// WithProductCategory 关联商品分类
func (r *productionRepo) WithProductCategory() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory")
	}
}

// WithProductCategoryMultiLanguageName 关联商品分类多语言
func (r *productionRepo) WithProductCategoryMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory.MultiLanguageName")
	}
}

// UpdateProduct 修改
func (r *productionRepo) UpdateProduct(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.ProductionOrderProduct{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

// UpdateOrder 修改
func (r *productionRepo) UpdateOrder(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.ProductionOrder{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

// GetProductsByPackageUuid 根据套餐SaleOrderProduct的uuid获取套餐下所有子商品的送厨单商品列表
func (r *productionRepo) GetProductsByPackageUuid(packageUuid uint64) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	err := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted).
		// r.db.Model(&model.SaleOrderProduct{}).Select("uuid").Where("package_uuid = ?" 表示获取套餐的子商品的sale_order_product_uuid列表
		// 然后通过子商品的sale_order_product_uuid获取生产订单商品列表
		Where("sale_order_product_uuid in (?)", r.db.Model(&model.SaleOrderProduct{}).Select("uuid").Where("package_uuid = ?", packageUuid).Scopes(NotDeleted)).
		Find(&productionOrderProducts).Error

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return productionOrderProducts, nil
}

// IsProductionFinishedBySaleBillUuid 检查销售账单下所有生产订单是否完成
func (r *productionRepo) IsProductionFinishedBySaleBillUuid(saleBillUuid uint64) (bool, error) {
	// 获取已叫production
	var count int64
	err := r.db.Model(&model.ProductionOrderProduct{}).Where("sale_bill_uuid = ? AND status < ?",
		saleBillUuid, constant.ProductionOrderProductStatusFinished).Count(&count).Error
	return count == 0, errors.WithMessage(err)
}

// 通过sale_bill_uuid和sale_order_product_uuid更新生产订单商品的batch_time、batch_tag_uuid
func (r *productionRepo) UpdateProductionOrderProductBatchTimeAndBatchTagUuid(saleBillUuid uint64, saleOrderProductUuids []uint64, batchTime int64, batchTagUuid uint64) error {
	return r.db.Model(&model.ProductionOrderProduct{}).Where("sale_bill_uuid = ? AND sale_order_product_uuid in (?)", saleBillUuid, saleOrderProductUuids).
		Update("batch_time", batchTime).Update("batch_tag_uuid", batchTagUuid).Error
}

// 非分批商品、或者分批已送厨商品
func (r *productionRepo) WhereIsNotBatchOrBatchTimeGT0() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_batch = 0 OR batch_time > 0")
	}
}

// GetProductionOrderList 获取生产订单列表
func (r *productionRepo) GetProductionOrderList(pageNo, pageSize int, productBomUuids []uint64, startTime int64, endTime int64, opts ...DBOption) ([]*model.ProductionOrderProduct, error) {
	var productionOrderProducts []*model.ProductionOrderProduct

	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if len(productBomUuids) > 0 {
		db = db.Where("product_bom_uuid in (?)", productBomUuids)
	}
	err := db.
		Where("status = ?", constant.ProductionOrderProductStatusFinished). // 已经完成出餐的商品
		Where("finished_time BETWEEN ? AND ?", startTime, endTime).         // 选择时间区间
		Order("finished_time desc").                                        // 按照完成时间最新的在前
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&productionOrderProducts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productionOrderProducts, nil
}

// GetProductionOrderList 获取生产订单列表
func (r *productionRepo) GetProductionOrderListCount(productBomUuids []uint64, startTime int64, endTime int64, opts ...DBOption) (int64, error) {
	var count int64

	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if len(productBomUuids) > 0 {
		db = db.Where("product_bom_uuid in (?)", productBomUuids)
	}
	err := db.
		Where("status = ?", constant.ProductionOrderProductStatusFinished). // 已经完成出餐的商品
		Where("finished_time BETWEEN ? AND ?", startTime, endTime).         // 选择时间区间
		Count(&count).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return count, nil
}

// GetProductsByUuids 根据Uuid列表获取生产订单商品
func (r *productionRepo) GetProductsByUuids(uuids []uint64, opts ...DBOption) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("uuid in (?)", uuids).Find(&productionOrderProducts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productionOrderProducts, nil
}
