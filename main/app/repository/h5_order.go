package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IH5OrderRepo 接单
type IH5OrderRepo interface {
	PaginateGetH5Order(pageNo, pageSize int, opts ...DBOption) ([]model.H5Order, int64, error)
	GetH5Order(opts ...DBOption) (model.H5Order, error)
	GetH5OrderUuids(opts ...DBOption) ([]uint64, error)
	GetH5OrderCount(opts ...DBOption) (int64, error)
	UpdateH5Order(qrcodeOrderUuid uint64, vars map[string]any) error

	CreateH5Order(qrcodeOrder model.H5Order) (uint64, error)
	DeleteH5Order(qrcodeOrderUuid uint64) error
	Reject(DeskUuid uint64) error

	WhereUuid(uuid uint64) DBOption           // 扫码订单uuid条件
	WhereUuidIn(uuids []uint64) DBOption      // 扫码订单uuid条件
	WhereStatus(status []uint) DBOption       // 扫码订单状态条件
	WhereNotStatus(status []uint) DBOption    // 扫码订单非状态
	WhereDeskRegionUuid(uuid uint64) DBOption // 扫码订单桌台区域id条件

	WithDesk() DBOption                                            // 关联桌台
	WithSaleOrderProducts() DBOption                               // 关联销售订单商品
	WithSaleOrderProductMultiLanguageName() DBOption               // 关联销售订单商品关联多语言
	WithH5OrderProducts() DBOption                                 // 关联扫码订单商品
	WithH5OrderProductSaleOrderProduct() DBOption                  // 关联扫码订单商品关联销售订单商品
	WithH5OrderProductSaleOrderProductMultiLanguageName() DBOption // 关联扫码订单商品关联销售订单商品关联多语言
	WithCashier() DBOption                                         // 关联收银员

	// 扫码订单商品相关

	GetH5OrderProducts(opts ...DBOption) ([]model.H5OrderProduct, error)                    // 扫码订单商品
	CreateH5OrderProduct(h5OrderProduct model.H5OrderProduct) (model.H5OrderProduct, error) // 快照销售订单商品

	WhereSaleBillUuid(uuid uint64) DBOption // 扫码订单商品销售账单Uuid条件

	WithSaleOrderProduct() DBOption // 关联销售订单商品
	WithH5Order() DBOption          // 关联扫码订单
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
		return nil, 0, err
	}
	// 获取分页数据
	err := db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&qrcodeOrders).Error
	return qrcodeOrders, total, err
}

// GetH5Order 获取接单详情
func (r *H5OrderRepoImpl) GetH5Order(opts ...DBOption) (model.H5Order, error) {
	var h5Order model.H5Order
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&h5Order).Error
	return h5Order, err
}

// GetH5OrderUuids 获取h5订单uuid
func (r *H5OrderRepoImpl) GetH5OrderUuids(opts ...DBOption) ([]uint64, error) {
	var h5OrderUuids []uint64
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("uuid").Scan(&h5OrderUuids).Error
	return h5OrderUuids, err
}

// GetH5OrderCount 获取H5订单数量
func (r *H5OrderRepoImpl) GetH5OrderCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.H5Order{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// UpdateH5Order 更新接单
func (r *H5OrderRepoImpl) UpdateH5Order(qrcodeOrderUuid uint64, vars map[string]any) error {
	if err := r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

// CreateH5Order 创建接单
func (r *H5OrderRepoImpl) CreateH5Order(qrcodeOrder model.H5Order) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&qrcodeOrder).Error; err != nil {
		return 0, err
	}
	return qrcodeOrder.Uuid, nil
}

// DeleteH5Order 软删除接单
func (r *H5OrderRepoImpl) DeleteH5Order(qrcodeOrderUuid uint64) error {
	return r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Update("delete_time", uint(time.Now().Unix())).Error
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

func (r *H5OrderRepoImpl) GetH5OrderProducts(opts ...DBOption) ([]model.H5OrderProduct, error) {
	var products []model.H5OrderProduct
	db := r.db.Model(&model.H5OrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&products).Error
	return products, err
}

func (r *H5OrderRepoImpl) CreateH5OrderProduct(h5OrderProduct model.H5OrderProduct) (model.H5OrderProduct, error) {
	err := r.db.Model(&model.H5OrderProduct{}).Create(&h5OrderProduct).Error
	return h5OrderProduct, err
}

func (r *H5OrderRepoImpl) WhereStatus(status []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status in (?)", status)
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

func (r *H5OrderRepoImpl) WithSaleOrderProducts() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProducts")
	}
}

func (r *H5OrderRepoImpl) WithSaleOrderProductMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProducts.MultiLanguageName")
	}
}

func (r *H5OrderRepoImpl) WithH5OrderProducts() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("H5OrderProducts")
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
		return db.Preload("Cashier")
	}
}
