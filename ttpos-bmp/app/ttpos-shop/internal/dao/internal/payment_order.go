// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentOrderDao is the data access object for the table ttpos_payment_order.
type PaymentOrderDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  PaymentOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// PaymentOrderColumns defines and stores column names for the table ttpos_payment_order.
type PaymentOrderColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 支付订单ID
	PaymentMethodName    string // 支付类型名称
	PaymentMethodUuid    string // 支付类型ID
	PaymentFeePercent    string // 支付手续费百分比,取值范围0-1
	RelatedType          string // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid          string // 关联的充值订单、销售订单ID
	CurrencyUnit         string // 货币单位
	PaymentAmount        string // 支付金额
	PaymentCommissionFee string // 支付手续费,支付金额*支付手续费百分比
	Amount               string // 实收金额，实收金额=支付金额+支付手续费
	TransactionNumber    string // 交易号
	Status               string // 支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常
	StatusReason         string // 支付状态原因
	BalanceAmount        string // 主账户金额,用于反结账时退款
	GiftBalanceAmount    string // 赠送帐户金额,用于反结账时退款
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// paymentOrderColumns holds the columns for the table ttpos_payment_order.
var paymentOrderColumns = PaymentOrderColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	PaymentMethodName:    "payment_method_name",
	PaymentMethodUuid:    "payment_method_uuid",
	PaymentFeePercent:    "payment_fee_percent",
	RelatedType:          "related_type",
	RelatedUuid:          "related_uuid",
	CurrencyUnit:         "currency_unit",
	PaymentAmount:        "payment_amount",
	PaymentCommissionFee: "payment_commission_fee",
	Amount:               "amount",
	TransactionNumber:    "transaction_number",
	Status:               "status",
	StatusReason:         "status_reason",
	BalanceAmount:        "balance_amount",
	GiftBalanceAmount:    "gift_balance_amount",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewPaymentOrderDao creates and returns a new DAO object for table data access.
func NewPaymentOrderDao(handlers ...gdb.ModelHandler) *PaymentOrderDao {
	return &PaymentOrderDao{
		group:    "default",
		table:    "ttpos_payment_order",
		columns:  paymentOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PaymentOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PaymentOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PaymentOrderDao) Columns() PaymentOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PaymentOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PaymentOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PaymentOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
