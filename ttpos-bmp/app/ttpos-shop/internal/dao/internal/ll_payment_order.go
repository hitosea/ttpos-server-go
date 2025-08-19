// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LlPaymentOrderDao is the data access object for the table ttpos_ll_payment_order.
type LlPaymentOrderDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  LlPaymentOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// LlPaymentOrderColumns defines and stores column names for the table ttpos_ll_payment_order.
type LlPaymentOrderColumns struct {
	Id                string // 自增ID
	Uuid              string // UUID
	PaymentOrderUuid  string // 自己系统的支付订单ID
	PaymentMethodUuid string // 支付方式ID
	RelatedType       string // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid       string // 关联的充值订单、销售订单ID
	MerchantId        string // lianlian商户号
	MerchantOrderId   string // 自己系统的为支付生成的订单号
	OrderId           string // lianlian订单ID
	OrderType         string // 订单类型
	OrderStatus       string // lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期
	OrderAmount       string // lianlian订单金额
	OrderCurrency     string // lianlian订单货币
	CommissionFee     string // 支付手续费,支付金额*支付手续费百分比
	FullName          string // 订单人名称
	OrderDesc         string // 订单描述
	LinkUrl           string // lianlian订单支付链接
	MerchantUserId    string // 自己系统的用户ID
	LlCreateTime      string // lianlian订单创建时间
	ExpiredTime       string // 过期时间
	PayTime           string // 支付时间
	CreateTime        string // 创建时间
	UpdateTime        string // 更新时间
	DeleteTime        string // 删除时间(时间戳)
}

// llPaymentOrderColumns holds the columns for the table ttpos_ll_payment_order.
var llPaymentOrderColumns = LlPaymentOrderColumns{
	Id:                "id",
	Uuid:              "uuid",
	PaymentOrderUuid:  "payment_order_uuid",
	PaymentMethodUuid: "payment_method_uuid",
	RelatedType:       "related_type",
	RelatedUuid:       "related_uuid",
	MerchantId:        "merchant_id",
	MerchantOrderId:   "merchant_order_id",
	OrderId:           "order_id",
	OrderType:         "order_type",
	OrderStatus:       "order_status",
	OrderAmount:       "order_amount",
	OrderCurrency:     "order_currency",
	CommissionFee:     "commission_fee",
	FullName:          "full_name",
	OrderDesc:         "order_desc",
	LinkUrl:           "link_url",
	MerchantUserId:    "merchant_user_id",
	LlCreateTime:      "ll_create_time",
	ExpiredTime:       "expired_time",
	PayTime:           "pay_time",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewLlPaymentOrderDao creates and returns a new DAO object for table data access.
func NewLlPaymentOrderDao(handlers ...gdb.ModelHandler) *LlPaymentOrderDao {
	return &LlPaymentOrderDao{
		group:    "default",
		table:    "ttpos_ll_payment_order",
		columns:  llPaymentOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LlPaymentOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LlPaymentOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LlPaymentOrderDao) Columns() LlPaymentOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LlPaymentOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LlPaymentOrderDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *LlPaymentOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
