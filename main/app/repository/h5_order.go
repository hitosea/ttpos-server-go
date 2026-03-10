package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IH5OrderRepo 接单
type IH5OrderRepo interface {
	IH5OrderQueryRepo
	Update(data map[string]interface{}, opts ...DBOption) error // 更新订单商品
	UpdateH5Order(qrcodeOrderUuid uint64, vars map[string]any) error
	UpdateH5OrderRecord(obj model.H5Order) error // 更新接单记录
	UpdateH5OrderProductRecord(obj model.H5OrderProduct) error

	CreateH5Order(qrcodeOrder model.H5Order) (uint64, error)
	DeleteH5Order(qrcodeOrderUuid uint64) error
	DeleteH5OrderProduct(h5OrderUuid, saleOrderProductUuid uint64) error
	Reject(DeskUuid uint64) error

	WhereUuid(uuid uint64) DBOption           // 扫码订单uuid条件
	WhereUuidIn(uuids []uint64) DBOption      // 扫码订单uuid条件
	WhereStatus(status []uint) DBOption       // 扫码订单状态条件
	WhereIsNeedAudit(flag int) DBOption       // 需要审核条件
	WhereUnNotified() DBOption                // 未处理的、 或已自动接单的h5订单
	WhereCreateTimeGt(ts int64) DBOption      // 创建时间大于
	WhereNotStatus(status []uint) DBOption    // 扫码订单非状态
	WhereDeskRegionUuid(uuid uint64) DBOption // 扫码订单桌台区域id条件

	WithDesk() DBOption                                            // 关联桌台
	WithSaleOrderProductsMultiLanguageName() DBOption              // 关联销售订单商品关联多语言
	WithSaleOrderProductMultiLanguageName() DBOption               // 关联销售订单商品关联多语言
	WithH5OrderProductSaleOrderProduct() DBOption                  // 关联扫码订单商品关联销售订单商品
	WithH5OrderProductSaleOrderProductMultiLanguageName() DBOption // 关联扫码订单商品关联销售订单商品关联多语言
	WithCashier() DBOption                                         // 关联收银员

	CreateH5OrderProduct(h5OrderProduct model.H5OrderProduct) (*model.H5OrderProduct, error) // 快照销售订单商品

	WhereSaleBillUuid(uuid uint64) DBOption // 扫码订单商品销售账单Uuid条件
	WhereOrderType(orderType int) DBOption  // 订单类型条件

	WithSaleOrderProduct() DBOption // 关联销售订单商品
	WithH5Order() DBOption          // 关联扫码订单
}

// IH5OrderQueryRepo 扫码订单查询
type IH5OrderQueryRepo interface {
	PaginateGetH5Order(pageNo, pageSize int, opts ...DBOption) ([]model.H5Order, int64, error)
	GetH5Order(opts ...DBOption) (*model.H5Order, error)
	GetH5OrderList(opts ...DBOption) ([]*model.H5Order, error)                  // 获取h5订单列表
	GetH5OrderProductCount(h5OrderUuid uint64) (int64, error)                   // 获取h5订单的商品数量
	GetH5OrderDetailOrdered(uuid uint64) (*model.H5Order, error)                // 获取接单详情，只获取已下单商品
	GetH5OrderListBySaleBillUuid(saleBillUuid uint64) ([]*model.H5Order, error) // 获取销售账单的所有待接单h5订单列表
	GetH5OrderUuids(opts ...DBOption) ([]uint64, error)
	GetH5OrderCount(opts ...DBOption) (int64, error)
	// 扫码订单商品相关
	GetH5OrderProducts(opts ...DBOption) ([]*model.H5OrderProduct, error)                           // 扫码订单商品
	GetH5OrderProductsBySaleBillUuidAndAccept(saleBillUuid uint64) ([]*model.H5OrderProduct, error) // 获取同一个销售账单，已接单的，h5订单商品
	GetH5OrderDetail(h5OrderUuid uint64, isNeedAudit bool) (*model.H5Order, error)                  // 扫码订单详情
}

func NewH5OrderRepo(db *gorm.DB) IH5OrderRepo {
	return NewH5OrderRepoImpl(db)
}

// NewH5OrderRepoImpl 创建新的仓库实现
func NewH5OrderRepoImpl(db *gorm.DB) *H5OrderRepoImpl {
	return &H5OrderRepoImpl{db: db}
}

type H5OrderRepoImpl struct {
	db *gorm.DB
}

// PaginateGetH5Order 分页获取接单列表
func (r *H5OrderRepoImpl) PaginateGetH5Order(pageNo, pageSize int, opts ...DBOption) ([]model.H5Order, int64, error) {
	var qrcodeOrders []model.H5Order
	var total int64
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取分页数据
	err := db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&qrcodeOrders).Error
	return qrcodeOrders, total, errors.WithMessage(err)
}

// GetH5Order 获取接单
func (r *H5OrderRepoImpl) GetH5Order(opts ...DBOption) (*model.H5Order, error) {
	var h5Order model.H5Order
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("create_time desc").First(&h5Order).Error
	return &h5Order, errors.WithMessage(err)
}

func (r *H5OrderRepoImpl) GetH5OrderDetailOrdered(uuid uint64) (*model.H5Order, error) {
	h5Order, err := r.GetH5Order(
		CommonRepo.WhereByUuid(uuid),
		r.WhereIsNeedAudit(1),
		r.WhereNotStatus([]uint{constant.H5OrderStatusChooseProduct}),
		CommonRepo.Preload(
			WithPreload{
				Query: "H5OrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "H5OrderProducts.SaleOrderProduct.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "Staff",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return h5Order, nil
}

// GetH5OrderList 获取h5订单列表
func (r *H5OrderRepoImpl) GetH5OrderList(opts ...DBOption) ([]*model.H5Order, error) {
	var h5Orders []*model.H5Order
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&h5Orders).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return h5Orders, nil
}

// GetH5OrderProductCount 获取h5订单的商品数量
func (r *H5OrderRepoImpl) GetH5OrderProductCount(h5OrderUuid uint64) (int64, error) {
	var count int64
	if err := r.db.Model(&model.H5OrderProduct{}).Where("h5_order_uuid = ? AND delete_time = ?", h5OrderUuid, constant.NotDeleted).Count(&count).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return count, nil
}

func (r *H5OrderRepoImpl) GetH5OrderListBySaleBillUuid(saleBillUuid uint64) ([]*model.H5Order, error) {
	h5Orders, err := r.GetH5OrderList(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		CommonRepo.WhereByStatus(constant.H5OrderStatusOrder),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return h5Orders, nil
}

// GetH5OrderDetail 获取接单详情
func (r *H5OrderRepoImpl) GetH5OrderDetail(h5OrderUuid uint64, isNeedAudit bool) (*model.H5Order, error) {
	needAudit := map[bool]DBOption{
		true:  r.WhereIsNeedAudit(1), // 查找需要审核
		false: r.WhereIsNeedAudit(0), // 查找不需要审核，自动接单
	}
	h5Order, err := r.GetH5Order(
		CommonRepo.WhereByUuid(h5OrderUuid),
		needAudit[isNeedAudit],
		CommonRepo.Preload(
			WithPreload{
				Query: "H5OrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "H5OrderProducts.SaleOrderProduct",
			},
			WithPreload{
				Query: "H5OrderProducts.SaleOrderProduct.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrderProducts",
			},
			WithPreload{
				Query: "SaleOrder.SaleBill",
			},
			// =================start 为了送厨检查 =================
			WithPreload{
				Query: "SaleOrderProducts",
			},
			WithPreload{
				Query: "SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrderProducts.ReturnOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrderProducts.CancelReasons",
			},
			WithPreload{
				Query: "SaleOrderProducts.ProductPackage",
			},
			WithPreload{
				Query: "SaleOrderProducts.ProductPackage.DineTax",
			},
			WithPreload{
				Query: "SaleOrderProducts.ProductPackage.TakeoutTax",
			},
			WithPreload{
				Query: "SaleOrderProducts.ProductPackage.ProductCategory",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductAttributes",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				// 用于检查商品包是否下架
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductPackage",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.FlavorMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.SauceMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			// =================end 为了送厨检查 =================
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return h5Order, nil
}

// GetH5OrderUuids 获取h5订单uuid
func (r *H5OrderRepoImpl) GetH5OrderUuids(opts ...DBOption) ([]uint64, error) {
	var h5OrderUuids []uint64
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("uuid").Scan(&h5OrderUuids).Error
	return h5OrderUuids, errors.WithMessage(err)
}

// GetH5OrderCount 获取H5订单数量
func (r *H5OrderRepoImpl) GetH5OrderCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return total, nil
}

// Update 更新
func (r *H5OrderRepoImpl) Update(data map[string]interface{}, opts ...DBOption) error {
	db := r.db.Model(&model.H5Order{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Updates(data).Error

	return errors.WithMessage(err)
}

// UpdateH5Order 更新接单
func (r *H5OrderRepoImpl) UpdateH5Order(qrcodeOrderUuid uint64, vars map[string]any) error {
	if err := r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Updates(vars).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateH5OrderRecord 更新接单记录
func (r *H5OrderRepoImpl) UpdateH5OrderRecord(obj model.H5Order) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.New("主键不能为空")
	}
	return r.db.Model(&model.H5Order{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
}

// UpdateH5OrderProductRecord 更新扫码订单商品记录
func (r *H5OrderRepoImpl) UpdateH5OrderProductRecord(obj model.H5OrderProduct) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.New("主键不能为空")
	}
	return r.db.Model(&model.H5OrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
}

// CreateH5Order 创建接单
func (r *H5OrderRepoImpl) CreateH5Order(obj model.H5Order) (uint64, error) {
	obj.SetNil()
	err := r.db.Model(&model.H5Order{}).Create(&obj).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return obj.Uuid, nil
}

// DeleteH5Order 软删除接单
func (r *H5OrderRepoImpl) DeleteH5Order(qrcodeOrderUuid uint64) error {
	return r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// DeleteH5OrderProduct 软删除订单商品。
func (r *H5OrderRepoImpl) DeleteH5OrderProduct(h5OrderUuid, saleOrderProductUuid uint64) error {
	return r.db.Model(&model.H5OrderProduct{}).Where("h5_order_uuid = ? AND sale_order_product_uuid = ?", h5OrderUuid, saleOrderProductUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// Reject 拒绝接单
func (r *H5OrderRepoImpl) Reject(DeskUuid uint64) error {
	return r.db.Model(&model.H5Order{}).
		Where("status = ?", 0).
		Where("desk_uuid = ?", DeskUuid).
		Updates(map[string]interface{}{
			"status":      2,
			"handle_time": uint(time.Now().Unix()),
		}).Error
}

func (r *H5OrderRepoImpl) GetH5OrderProducts(opts ...DBOption) ([]*model.H5OrderProduct, error) {
	var products []*model.H5OrderProduct
	db := r.db.Model(&model.H5OrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&products).Error
	return products, errors.WithMessage(err)
}

// GetH5OrderProductsBySaleBillUuidAndAccept 获取同一个销售账单，已接单的，h5订单商品
func (r *H5OrderRepoImpl) GetH5OrderProductsBySaleBillUuidAndAccept(saleBillUuid uint64) ([]*model.H5OrderProduct, error) {
	products, err := r.GetH5OrderProducts(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "H5Order",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereByStatus(constant.H5OrderStatusAccepted)),
				},
			},
			WithPreload{
				Query: "SaleOrderProduct.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	newProducts := make([]*model.H5OrderProduct, 0)
	for _, product := range products {
		if product.IsAccepted() {
			newProducts = append(newProducts, product)
		}
	}
	return newProducts, nil
}

func (r *H5OrderRepoImpl) CreateH5OrderProduct(h5OrderProduct model.H5OrderProduct) (*model.H5OrderProduct, error) {
	h5OrderProduct.SetNil()
	err := r.db.Model(&model.H5OrderProduct{}).Create(&h5OrderProduct).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &h5OrderProduct, nil
}

func (r *H5OrderRepoImpl) WhereStatus(status []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status in (?)", status)
	}
}

func (r *H5OrderRepoImpl) WhereIsNeedAudit(flag int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_need_audit = ?", flag)
	}
}

func (r *H5OrderRepoImpl) WhereUnNotified() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ? OR (status = ? AND is_auto_accept = 1)", constant.H5OrderStatusOrder, constant.H5OrderStatusAccepted)
	}
}

func (r *H5OrderRepoImpl) WhereCreateTimeGt(ts int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time > ?", ts)
	}
}

func (r *H5OrderRepoImpl) WhereNotStatus(status []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status not in (?)", status)
	}
}

func (r *H5OrderRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *H5OrderRepoImpl) WhereUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid in (?)", uuids)
	}
}

func (r *H5OrderRepoImpl) WhereDeskRegionUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("desk_uuid in (?)", r.db.Model(&model.Desk{}).Select("uuid").Where("region_uuid = ?", uuid))
	}
}

func (r *H5OrderRepoImpl) WhereSaleBillUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid = ?", uuid)
	}
}

func (r *H5OrderRepoImpl) WhereOrderType(orderType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if orderType < 0 {
			return db // -1 表示全部，不添加条件
		}
		return db.Where("order_type = ?", orderType)
	}
}

func (r *H5OrderRepoImpl) WithSaleOrderProduct() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProduct")
	}
}
func (r *H5OrderRepoImpl) WithH5Order() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("H5Order")
	}
}

func (r *H5OrderRepoImpl) WithDesk() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Desk")
	}
}

func (r *H5OrderRepoImpl) WithSaleOrderProductsMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProducts.MultiLanguageName")
	}
}

func (r *H5OrderRepoImpl) WithSaleOrderProductMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProduct.MultiLanguageName")
	}
}

func (r *H5OrderRepoImpl) WithH5OrderProductSaleOrderProduct() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("H5OrderProducts.SaleOrderProduct")
	}
}

func (r *H5OrderRepoImpl) WithH5OrderProductSaleOrderProductMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("H5OrderProducts.SaleOrderProduct.MultiLanguageName")
	}
}

func (r *H5OrderRepoImpl) WithCashier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Staff")
	}
}
