package repository

import (
	"strings"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberRechargeOrderRepo interface {
	IMemberRechargeOrderQueryRepo
	WithMember() DBOption                                                                           // 预加载充值会员
	WithStaff() DBOption                                                                            // 预加载收银员
	WithPaymentOrders() DBOption                                                                    // 预加载支付订单
	WithPaymentOrderPaymentMethod() DBOption                                                        // 预加载支付订单.支付方式
	WithPaymentOrderReturnOrderAmount() DBOption                                                    // 预加载支付订单.退款金额
	WithRechargeOrderOperationLogs() DBOption                                                       // 预加载操作日志
	WithReturnOrders() DBOption                                                                     // 关联退货单
	WithReturnOrderAmount() DBOption                                                                // 关联退货单.退款金额
	WhereUuid(uuid uint64) DBOption                                                                 // Uuid条件
	WhereStatus(status int) DBOption                                                                // 状态条件
	WhereOrderNoLike(keyword string) DBOption                                                       // 订单编号模糊搜索
	WhereCreateTimeBetween(start, end int64) DBOption                                               // 创建时间范围
	WhereTimeBetween(params TimeQueryParams) DBOption                                               // 多个时间范围
	Create(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error)              // 创建充值订单
	CreateOldRecord(rechargeOrder model.MemberRechargeOrder) error                                  // 创建充值订单,仅用于迁移数据
	Update(uuid uint64, vars map[string]any) error                                                  // 更新充值订单
	UpdateErpProductsInvoiceName(uuid uint64, erpProductsInvoiceName string) error                  // 更新充值订单的商品发票名称
	UpdateErpSalesInvoiceInfo(uuid uint64, salesInvoiceName string, paymentEntryNames string) error // 更新充值订单的 SI 名称和 PE 名称
}

// IMemberRechargeOrderQueryRepo 充值订单查询
type IMemberRechargeOrderQueryRepo interface {
	GetRechargeOrder(opts ...DBOption) model.MemberRechargeOrder                                                     // 获取充值订单
	GetOrderCount(opts ...DBOption) (int64, error)                                                                   // 获取订单数量
	PaginateGetRechargeOrder(pageNo int, pageSize int, opts ...DBOption) ([]model.MemberRechargeOrder, int64, error) // 分页获取充值订单
	GetRechargeOrderByUuid(uuid uint64) (model.MemberRechargeOrder, error)                                           // 通过Uuid获取充值订单
	GetRechargeOrderAllInfo(uuid uint64) model.MemberRechargeOrder                                                   // 获取充值订单所有信息
	GetMemberRechargeOrderList(opts ...DBOption) ([]model.MemberRechargeOrder, error)                                // 获取充值订单列表
}

func NewMemberRechargeOrderRepo(db *gorm.DB) IMemberRechargeOrderRepo {
	return NewMemberRechargeOrderRepoImpl(db)
}

type memberRechargeOrderRepo struct {
	db *gorm.DB
}

func NewMemberRechargeOrderRepoImpl(db *gorm.DB) IMemberRechargeOrderRepo {
	return &memberRechargeOrderRepo{db: db}
}

// GetRechargeOrder 获取充值订单
func (r *memberRechargeOrderRepo) GetRechargeOrder(opts ...DBOption) model.MemberRechargeOrder {
	var rechargeOrder model.MemberRechargeOrder
	db := r.db.Model(&model.MemberRechargeOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&rechargeOrder)
	return rechargeOrder
}

// GetOrderCount 获取充值订单数量
func (r *memberRechargeOrderRepo) GetOrderCount(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.MemberRechargeOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&count).Error
	return count, errors.WithMessage(err)
}

// PaginateGetRechargeOrder 分页获取充值订单
func (r *memberRechargeOrderRepo) PaginateGetRechargeOrder(pageNo int, pageSize int, opts ...DBOption) ([]model.MemberRechargeOrder, int64, error) {
	var rechargeOrders []model.MemberRechargeOrder
	var total int64
	db := r.db.Model(&model.MemberRechargeOrder{}).Session(&gorm.Session{})
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("create_time desc").Find(&rechargeOrders).Error
	return rechargeOrders, total, errors.WithMessage(err)
}

// WithPaymentOrders 预加载支付订单
func (r *memberRechargeOrderRepo) WithPaymentOrders() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted).Where("related_type = 1") // 关联支付订单类型为充值订单
		})
	}
}

// WithRefundOrders 预加载退款订单
func (r *memberRechargeOrderRepo) WithRefundOrders() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("RefundOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted).Where("related_order_type = 1") // 关联退款订单类型为充值订单
		})
	}
}

// WithMember 预加载充值会员
func (r *memberRechargeOrderRepo) WithMember() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Member")
	}
}

// WithStaff 预加载收银员
func (r *memberRechargeOrderRepo) WithStaff() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Staff")
	}
}

// WhereUuid uuid 条件
func (r *memberRechargeOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereStatus status 条件
func (r *memberRechargeOrderRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereOrderNoLike order_no 模糊搜索
func (r *memberRechargeOrderRepo) WhereOrderNoLike(keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_no like ?", "%"+keyword+"%")
	}
}

// WhereCreateTimeBetween 创建时间范围
func (r *memberRechargeOrderRepo) WhereCreateTimeBetween(start, end int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time BETWEEN ? AND ?", start, end)
	}
}

type TimeRange struct {
	Field     string // 字段
	StartTime int64  // 开始时间
	EndTime   int64  // 结束时间
}
type TimeQueryParams struct {
	TimeRanges []TimeRange // 时间字段范围
	Operator   string      // 可以是 OR 或者 AND
}

// WhereTimeBetween 时间范围
func (r *memberRechargeOrderRepo) WhereTimeBetween(params TimeQueryParams) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		var queries []string
		var query string
		var args []any
		for _, timeRange := range params.TimeRanges {
			var singleQueries []string
			if timeRange.StartTime != 0 {
				singleQueries = append(singleQueries, timeRange.Field+" >= ?")
				args = append(args, timeRange.StartTime)
			}
			if timeRange.EndTime != 0 {
				singleQueries = append(singleQueries, timeRange.Field+" <= ?")
				args = append(args, timeRange.EndTime)
			}
			queries = append(queries, strings.Join(singleQueries, " AND "))
		}
		switch params.Operator {
		case "OR":
			query = strings.Join(queries, " OR ")
		case "AND":
			query = strings.Join(queries, " AND ")
		default:
			return db
		}
		return db.Where(query, args...)
	}
}

// WithPaymentOrderPaymentMethod 预加载支付订单.支付方式
func (r *memberRechargeOrderRepo) WithPaymentOrderPaymentMethod() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders.PaymentMethod")
	}
}

// WithPaymentOrderReturnOrderAmount 预加载支付订单.退款金额
func (r *memberRechargeOrderRepo) WithPaymentOrderReturnOrderAmount() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders.ReturnOrderAmounts")
	}
}

// WithRechargeOrderOperationLogs 预加载操作日志
func (r *memberRechargeOrderRepo) WithRechargeOrderOperationLogs() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("RechargeOrderOperationLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time desc")
		})
	}
}

// WithReturnOrders 预加退货订单
func (r *memberRechargeOrderRepo) WithReturnOrders() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ReturnOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted).Where("related_order_type = 1").Where("is_reverse_settlement = 0")
		})
	}
}

// WithReturnOrderAmount 预加退货订单.退款金额
func (r *memberRechargeOrderRepo) WithReturnOrderAmount() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ReturnOrders.ReturnOrderAmounts", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// Update 修改充值订单
func (r *memberRechargeOrderRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// Create 创建充值订单
func (r *memberRechargeOrderRepo) Create(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error) {
	err := r.db.Model(&model.MemberRechargeOrder{}).Create(&rechargeOrder).Error
	return rechargeOrder, errors.WithMessage(err)
}

// CreateOldRecord 创建充值订单,仅用于迁移数据
func (r *memberRechargeOrderRepo) CreateOldRecord(obj model.MemberRechargeOrder) error {
	rechargeOrder := obj
	rechargeOrder.SetNil()
	if err := r.db.Model(&model.MemberRechargeOrder{}).Create(&rechargeOrder).Error; err != nil {
		return errors.WithMessage(err)
	}

	// 创建支付订单
	for _, paymentOrder := range obj.PaymentOrders {
		paymentOrder.SetNil()
		if err := r.db.Model(&model.PaymentOrder{}).Create(&paymentOrder).Error; err != nil {
			return errors.WithMessage(err)
		}
	}

	// 创建操作日志
	for _, operationLog := range obj.RechargeOrderOperationLogs {
		operationLog.SetNil()
		if err := r.db.Model(&model.MemberRechargeOrderOperationLog{}).Create(&operationLog).Error; err != nil {
			return errors.WithMessage(err)
		}
	}

	// 创建退货单

	for _, returnOrder := range obj.ReturnOrders {
		returnOrder.SetNil()
		if err := r.db.Model(&model.ReturnOrder{}).Create(&returnOrder).Error; err != nil {
			return errors.WithMessage(err)
		}
	}

	// 创建退货单金额
	for _, returnOrder := range obj.ReturnOrders {
		for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
			returnOrderAmount.SetNil()
			if err := r.db.Model(&model.ReturnOrderAmount{}).Create(&returnOrderAmount).Error; err != nil {
				return errors.WithMessage(err)
			}
		}
	}
	return nil
}

// GetRechargeOrderByUuid 通过Uuid获取充值订单
func (r *memberRechargeOrderRepo) GetRechargeOrderByUuid(uuid uint64) (model.MemberRechargeOrder, error) {
	var rechargeOrder model.MemberRechargeOrder
	err := r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).First(&rechargeOrder).Error
	return rechargeOrder, errors.WithMessage(err)
}

// GetRechargeOrderAllInfo 获取充值订单所有信息
func (r *memberRechargeOrderRepo) GetRechargeOrderAllInfo(uuid uint64) model.MemberRechargeOrder {
	rechargeOrder := r.GetRechargeOrder(
		r.WithPaymentOrders(),
		r.WithPaymentOrderPaymentMethod(),
		r.WithPaymentOrderReturnOrderAmount(),
		r.WithReturnOrders(),
		r.WithReturnOrderAmount(),
		r.WhereUuid(uuid),
	)

	return rechargeOrder
}

// GetMemberRechargeOrderList 获取充值订单列表
func (r *memberRechargeOrderRepo) GetMemberRechargeOrderList(opts ...DBOption) ([]model.MemberRechargeOrder, error) {
	var rechargeOrders []model.MemberRechargeOrder
	db := r.db.Model(&model.MemberRechargeOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&rechargeOrders).Error
	return rechargeOrders, errors.WithMessage(err)
}

// UpdateErpProductsInvoiceName 更新充值订单的商品发票名称
func (r *memberRechargeOrderRepo) UpdateErpProductsInvoiceName(uuid uint64, erpProductsInvoiceName string) error {
	return r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"erp_products_invoice_name": erpProductsInvoiceName,
	}).Error
}

// UpdateErpSalesInvoiceInfo 更新充值订单的 SI 名称和 PE 名称（MQ 回调使用）
func (r *memberRechargeOrderRepo) UpdateErpSalesInvoiceInfo(uuid uint64, salesInvoiceName string, paymentEntryNames string) error {
	return r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"erp_products_invoice_name": salesInvoiceName,
		"erp_payment_entry_names":   paymentEntryNames,
	}).Error
}
